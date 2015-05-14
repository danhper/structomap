# Go serializer [![Build Status](https://travis-ci.org/tuvistavie/serializer.svg)](https://travis-ci.org/tuvistavie/serializer) [![GoDoc](https://godoc.org/github.com/tuvistavie/serializer?status.svg)](https://godoc.org/github.com/tuvistavie/serializer)

This package helps you to serialize your `struct` into `map` easily. It provides a `serializer.Serializer` interface implemented by the `serializer.Base` type which contains chainable function to add, remove or modify fields. The result is returned as a `map[string]interface{}`.
It is then up to you to encode the result in JSON, XML or whatever you like.

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
           UseSnakeCase().
           Pick("ID", "FirstName", "LastName", "Email").
           OmitIf(func(u interface{}) bool {
               return u.(User).HideEmail
           }, "Email").
           Add("CurrentTime", time.Now()).
           AddFunc("FullName", func(u interface{}) interface{} {
               return u.(User).FirstName + " " + u.(User).LastName
           }).
           Result()
// -> {"id": 1, "first_name": "Foo", "last_name": "Bar", "current_time": time.Time{...}, "full_name": "Foo Bar"}
```

The full documentation is available at https://godoc.org/github.com/tuvistavie/serializer.

## Building your own serializer

With `Serializer` as a base, you can easily build your serializer.

```go
type UserSerializer struct {
  *serializer.Base
}

func NewUserSerializer(user User) *UserSerializer {
  u := &UserSerializer{serializer.New(user)}
  u.Pick("ID", "CreatedAt", "UpdatedAt", "DeletedAt")
  return u
}

func (u *UserSerializer) WithPrivateInfo() *UserSerializer {
  u.Pick("Email")
  return u
}

userMap := NewUserSerializer(user).WithPrivateInfo().Result()
```

Note that the `u.Pick`, and all other methods do modify the serializer, they do not return a new serializer each time. This is why it works
even when ignoring `u.Pick` return value.

## License

This is released under the MIT license. See the [LICENSE](./LICENSE) file for more information.
