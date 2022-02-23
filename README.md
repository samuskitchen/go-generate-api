# GO-GENERATE-API
go-generate-api is a library to which we provide an object and it creates the main CRUD operations.

## Dependencies:
- Gorm ORM library
- Echo Router

## Operations
- Create: `POST`
    - `localhost:8080/basepath`

- Find All: `GET`
    - `localhost:8080/basepath`

- Find by Identifier: `GET`
    - `localhost:8080/basepath/:identifier`

- UPDATE: `PUT`   : _id o identifier is necessary_
    - `localhost:8080/basepath`

- DELETE: `DELETE`
    - `localhost:8080/basepath/:identifier`


## Example
```go
package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/samuskitchen/go-generate-api"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Person struct {
	ID        uint   `json:"id"`
	FirstName string `json:"first_name"`
	Age       uint   `json:"age"`
}

type Product struct {
	Code  string  `json:"code"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func main() {
	dsn := "host=localhost user=postgres password=admin dbname=test port=5432 sslmode=disable"
	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true, 
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	
	if err != nil {
		log.Fatalln(err)
	}
	
	log.Println("Database opened!!")
	e := echo.New()

	h := generate.NewHandlerGenerate(e.Group("/person"), conn)
	h.Start(Person{})

	handlerProduct := generate.NewHandlerGenerate(e.Group("/product"), conn) 
	//Name of the primary key field, in the table and the model
	handlerProduct.Start(Product{}, generate.WithKeyFieldName("code", "Code", false))
	
	e.Start("localhost:8080")
}
```