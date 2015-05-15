package serializer

import (
	"encoding/json"
	"fmt"
	"github.com/tuvistavie/testify/assert"
	"testing"
	"time"
)

func alwaysTrue(u interface{}) bool {
	return true
}

func alwaysFalse(u interface{}) bool {
	return false
}

func identity(u interface{}) interface{} {
	return u
}

type User struct {
	ID        int
	Email     string
	Birthday  time.Time
	Age       int
	HideEmail bool
	FirstName string
	LastName  string
	HideName  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CustomSerializer struct {
	*Base
}

func NewCustomSerializer(entity interface{}) *CustomSerializer {
	return &CustomSerializer{Base: New(entity)}
}

func (c *CustomSerializer) WithBasicInfo() *CustomSerializer {
	c.Pick("ID", "FirstName", "LastName")
	return c
}

func (c *CustomSerializer) WithPrivateinfo() *CustomSerializer {
	c.Pick("Email")
	return c
}

var user = User{
	ID:        1,
	Email:     "x@example.com",
	Birthday:  time.Date(1989, 11, 24, 0, 0, 0, 0, time.UTC),
	Age:       25,
	FirstName: "Foo",
	LastName:  "Bar",
	HideEmail: true,
	HideName:  true,
	CreatedAt: time.Date(2015, 05, 13, 15, 30, 0, 0, time.UTC),
	UpdatedAt: time.Date(2015, 05, 13, 15, 30, 0, 0, time.UTC),
}

func ExampleSerializer() {
	userMap := New(user).
		UseSnakeCase().
		Pick("ID", "FirstName", "LastName", "Email").
		PickFunc(func(t interface{}) interface{} {
		return t.(time.Time).Format(time.RFC3339)
	}, "CreatedAt", "UpdatedAt").
		OmitIf(func(u interface{}) bool {
		return u.(User).HideEmail
	}, "Email").
		Add("CurrentTime", time.Date(2015, 5, 15, 17, 41, 0, 0, time.UTC)).
		AddFunc("FullName", func(u interface{}) interface{} {
		return u.(User).FirstName + " " + u.(User).LastName
	}).Result()
	str, _ := json.MarshalIndent(userMap, "", "  ")
	fmt.Println(string(str))
	// Output:
	// {
	//   "created_at": "2015-05-13T15:30:00Z",
	//   "current_time": "2015-05-15T17:41:00Z",
	//   "first_name": "Foo",
	//   "full_name": "Foo Bar",
	//   "id": 1,
	//   "last_name": "Bar",
	//   "updated_at": "2015-05-13T15:30:00Z"
	// }
}

func TestPickAll(t *testing.T) {
	m := New(user).PickAll().Result()
	assert.Contains(t, m, "ID")
	assert.Contains(t, m, "Age")
	assert.Contains(t, m, "FirstName")
}

func TestPick(t *testing.T) {
	m := New(user).Pick("ID", "Age").Result()
	assert.Contains(t, m, "ID")
	assert.Contains(t, m, "Age")
	assert.NotContains(t, m, "FirstName")
	m = New(user).Pick("ID").Pick("Age").Result()
	assert.Contains(t, m, "ID")
	assert.Contains(t, m, "Age")
	assert.Equal(t, m["ID"], user.ID)
}

func TestPickIf(t *testing.T) {
	m := New(user).
		PickIf(alwaysTrue, "ID", "FirstName").
		PickIf(alwaysFalse, "Email").Result()
	assert.Contains(t, m, "ID")
	assert.Contains(t, m, "FirstName")
	assert.NotContains(t, m, "Email")
	assert.NotContains(t, m, "Age")
}

func TestPickFunc(t *testing.T) {
	m := New(user).PickFunc(func(t interface{}) interface{} {
		return t.(time.Time).Format(time.RFC3339)
	}, "CreatedAt", "UpdatedAt").Result()
	assert.Contains(t, m, "CreatedAt")
	assert.Contains(t, m, "UpdatedAt")
	assert.Equal(t, m["CreatedAt"], user.CreatedAt.Format(time.RFC3339))
}

func TestPickFuncIf(t *testing.T) {
	m := New(user).PickFuncIf(alwaysTrue, func(t interface{}) interface{} {
		return t.(time.Time).Format(time.RFC3339)
	}, "CreatedAt", "UpdatedAt").PickFuncIf(alwaysFalse, identity, "Email").Result()
	assert.Contains(t, m, "CreatedAt")
	assert.Contains(t, m, "UpdatedAt")
	assert.NotContains(t, m, "Email")
	assert.Equal(t, m["CreatedAt"], user.CreatedAt.Format(time.RFC3339))
}

func TestMultipleOmit(t *testing.T) {
	m := New(user).PickAll().Omit("Birthday", "FirstName").Result()
	assert.NotContains(t, m, "Birthday")
	assert.NotContains(t, m, "FirstName")
	assert.Contains(t, m, "ID")
}

func TestOmitIf(t *testing.T) {
	m := New(user).PickAll().OmitIf(func(u interface{}) bool {
		return u.(User).HideName
	}, "FirstName", "LastName").Result()
	assert.NotContains(t, m, "FirstName", "LastName")

	m = New(user).PickAll().OmitIf(func(u interface{}) bool {
		return u.(User).Age < 18
	}, "Birthday", "Age").Result()
	assert.Contains(t, m, "Birthday")
	assert.Contains(t, m, "Age")
}

func TestAdd(t *testing.T) {
	m := New(user).Add("Foo", "Bar").Result()
	assert.Contains(t, m, "Foo")
	assert.Equal(t, "Bar", m["Foo"])
}

func TestAddIf(t *testing.T) {
	m := New(user).AddIf(alwaysFalse, "Foo", "Bar").Result()
	assert.NotContains(t, m, "Foo")
	m = New(user).AddIf(alwaysTrue, "Foo", "Bar").Result()
	assert.Contains(t, m, "Foo")
	assert.Equal(t, "Bar", m["Foo"])
}

func TestAddFunc(t *testing.T) {
	m := New(user).AddFunc("FullName", func(u interface{}) interface{} {
		return u.(User).FirstName + " " + u.(User).LastName
	}).Result()
	assert.Contains(t, m, "FullName")
	assert.Equal(t, user.FirstName+" "+user.LastName, m["FullName"])
}

func TestAddFuncIf(t *testing.T) {
	m := New(user).AddFuncIf(alwaysTrue, "Foo", func(u interface{}) interface{} {
		return "Bar"
	}).Result()
	assert.Contains(t, m, "Foo")
	assert.Equal(t, m["Foo"], "Bar")
	m = New(user).AddFuncIf(alwaysFalse, "Foo", identity).Result()
	assert.NotContains(t, m, "Foo")
}

func TestTransformKeys(t *testing.T) {
	m := New(user).PickAll().TransformKeys(func(s string) string {
		return "dummy_" + s
	}).Result()
	assert.Contains(t, m, "dummy_ID")
	assert.Contains(t, m, "dummy_FirstName")
	assert.NotContains(t, m, "ID")
}

func TestUseSnakeCase(t *testing.T) {
	m := New(user).UseSnakeCase().PickAll().Result()
	assert.Contains(t, m, "id")
	assert.Contains(t, m, "first_name")
	assert.NotContains(t, m, "FirstName")
}

func TestUseCamelCase(t *testing.T) {
	m := New(user).UseCamelCase().PickAll().Result()
	assert.Contains(t, m, "id")
	assert.Contains(t, m, "firstName")
	assert.NotContains(t, m, "FirstName")
}

func TestUsePascalCase(t *testing.T) {
	m := New(user).UsePascalCase().PickAll().Result()
	assert.Contains(t, m, "ID")
	assert.Contains(t, m, "FirstName")
}

func TestDefaultCase(t *testing.T) {
	SetDefaultCase(SnakeCase)
	m := New(user).PickAll().Result()
	assert.Contains(t, m, "first_name")
	SetDefaultCase(CamelCase)
	m = New(user).PickAll().Result()
	assert.Contains(t, m, "firstName")
	SetDefaultCase(PascalCase)
	m = New(user).PickAll().Result()
	assert.Contains(t, m, "FirstName")
}

func TestCustomSerializer(t *testing.T) {
	m := NewCustomSerializer(user).WithPrivateinfo().WithBasicInfo().Result()
	for _, field := range []string{"ID", "FirstName", "LastName", "Email"} {
		assert.Contains(t, m, field)
	}
}
