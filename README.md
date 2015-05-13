# Go serializer [![Build Status](https://travis-ci.org/tuvistavie/serializer.svg)](https://travis-ci.org/tuvistavie/serializer) [![GoDoc](https://godoc.org/github.com/tuvistavie/serializer?status.svg)](https://godoc.org/github.com/tuvistavie/serializer)

Serialize your `struct` easily.

Here is an example.

```go
import "github.com/tuvistavie/serializer"

type User struct {
    ID        int
    FirstName string
    LastName  string
    Email     string
    HideEmail bool
    CreatedAt time.Time
    UpdatedAt time.Time
}

user := &User{"ID": 1, "FirstName": "Foo", "LastName": "Bar", "Email": "foo@example.org", "HideEmail": true}

serializer.New(user).
           Pick("ID", "FirstName", "LastName", "Email").
           OmitIf(func(u interface{}) bool {
               return u.(User).HideEmail
           }, "Email").
           Add("CurrentTime", time.Now()).
           AddFunc("FullName", func(u interface{}) interface{} {
               return u.(User).FirstName + " " + u.(User).LastName
           }).
           Result()
// -> {"ID": 1, "FirstName": "Foo", "LastName": "Bar", "CurrentTime": time.Time{...}, "FullName": "Foo Bar"}
```
