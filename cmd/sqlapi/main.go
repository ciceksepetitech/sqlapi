package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/ciceksepetitech/sqlapi/internal/mongodb"
	"github.com/ciceksepetitech/sqlapi/internal/mssql"
	"github.com/ciceksepetitech/sqlapi/internal/mysql"
)

func main() {
	http.HandleFunc("/sql", query)

	log.Printf("API server listening at: 127.0.0.1:8033")
	err := http.ListenAndServe(":8033", nil)
	if err != nil {
		panic(err)
	}
}

func query(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	p := getPayload(r)

	var result []map[string]interface{}

	switch p.DB.Type {
	case "mysql":
		rows, err := mysql.I(p.DB.User, p.DB.Password, p.DB.Host, p.DB.Name).Query(p.Query)
		if err != nil {
			panic(err)
		}
		defer rows.Close()
		result = mapToInterface(rows)
		break
	case "mssql":
		rows, err := mssql.I(p.DB.User, p.DB.Password, p.DB.Host, p.DB.Name).Query(p.Query)
		if err != nil {
			panic(err)
		}
		defer rows.Close()
		result = mapToInterface(rows)
		break
	case "mongodb":
		var filter map[string]interface{}
		if err := json.Unmarshal([]byte(p.Query), &filter); err != nil {
			panic(err)
		}
		if v, ok := filter["_id"]; ok {
       		if primitive.IsValidObjectID(v.(string)){
            	id, err := primitive.ObjectIDFromHex(v.(string))
        		if err != nil {
        			panic(err)
        		}    	
				filter["_id"] = id  
        	} else{
            	filter["_id"] = v.(string)  
        	}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		client := mongodb.I(p.DB.User, p.DB.Password, p.DB.Host)
		cur, err := client.Database(p.DB.Name).Collection(p.Collection).Find(ctx, filter)
		if err != nil {
			panic(err)
		}
		defer cur.Close(ctx)
		result = mapToInterfaceMongo(ctx, cur)
		break
	default:
		panic("The db type you specified does not implemented.")
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(result)
}

func mapToInterfaceMongo(ctx context.Context, cur *mongo.Cursor) (result []map[string]interface{}) {
	for cur.Next(ctx) {
		var r bson.M
		err := cur.Decode(&r)
		if err != nil {
			panic(err)
		}
		result = append(result, r)
	}
	if err := cur.Err(); err != nil {
		panic(err)
	}
	return result
}

func mapToInterface(rows *sql.Rows) []map[string]interface{} {
	cols, _ := rows.Columns()

	values := make([][]byte, len(cols))
	var data []map[string]interface{}
	dest := make([]interface{}, len(cols)) // A temporary interface{} slice

	for i := range values {
		dest[i] = &values[i] // Put pointers to each string in the interface slice
	}

	for rows.Next() {
		result := make(map[string]interface{})
		rows.Scan(dest...)
		for i, key := range cols {
			val := values[i]
			if val == nil {
				result[key] = nil
			} else {
				var f interface{}
				if err := json.Unmarshal(val, &f); err == nil {
					result[key] = f
				} else if len(val) == 1 && val[0] <= 1 { // boolean
					result[key] = val[0] == 1
				} else {
					result[key] = string(val)
				}
			}
		}
		data = append(data, result)
	}

	return data
}

// SQLRequest .
type SQLRequest struct {
	Query      string `json:"query"`
	Collection string `json:"collection"`
	DB         *DB    `json:"db"`
}

// DB .
type DB struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Name     string `json:"name"`
	User     string `json:"user"`
	Password string `json:"password"`
}

func getPayload(r *http.Request) *SQLRequest {
	// Read body
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		panic(fmt.Sprintf("Invalid payload"))
	}

	// Unmarshal
	payload := &SQLRequest{}
	err = json.Unmarshal(body, &payload)
	if err != nil {
		panic(err)
	}

	return payload
}
