package serializer

import (
	"github.com/tuvistavie/testify/assert"
	"strings"
	"testing"
	"time"
)

type User struct {
	ID        int
	Email     string
	Birthday  time.Time
	Age       int
	FirstName string
	LastName  string
	Num       int
	HideName  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	BillingAddressID int
	IgnoreMe         int
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
	Birthday:  time.Date(1989, 11, 24, 0, 0, 0, 0, time.UTC),
	Age:       25,
	FirstName: "Foo",
	LastName:  "Bar",
	Num:       8,
	HideName:  true,
	CreatedAt: time.Date(2015, 05, 13, 15, 30, 0, 0, time.UTC),
	UpdatedAt: time.Date(2015, 05, 13, 15, 30, 0, 0, time.UTC),
	DeletedAt: nil,
}

func TestPickAll(t *testing.T) {
	m := New(user).PickAll().Result()
	assert.Contains(t, m, "ID")
	assert.Contains(t, m, "Age")
	assert.Contains(t, m, "FirstName")
	assert.Contains(t, m, "Num")
}

func TestPick(t *testing.T) {
	m := New(user).Pick("ID", "Age").Result()
	assert.Contains(t, m, "ID")
	assert.Contains(t, m, "Age")
	assert.NotContains(t, m, "FirstName")
	assert.NotContains(t, m, "Num")
	m = New(user).Pick("ID").Pick("Age").Result()
	assert.Contains(t, m, "ID")
	assert.Contains(t, m, "Age")
	assert.Equal(t, m["ID"], user.ID)
}

func TestPickIf(t *testing.T) {
	m := New(user).PickIf(func(u interface{}) bool {
		return true
	}, "ID", "FirstName").PickIf(func(u interface{}) bool {
		return false
	}, "Email").Result()
	assert.Contains(t, m, "ID")
	assert.Contains(t, m, "FirstName")
	assert.NotContains(t, m, "Email")
	assert.NotContains(t, m, "Age")
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
	m := New(user).AddIf(func(u interface{}) bool {
		return false
	}, "Foo", "Bar").Result()
	assert.NotContains(t, m, "Foo")
	m = New(user).AddIf(func(u interface{}) bool {
		return true
	}, "Foo", "Bar").Result()
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
	m := New(user).AddFuncIf(func(u interface{}) bool {
		return true
	}, "Foo", func(u interface{}) interface{} {
		return "Bar"
	}).Result()
	assert.Contains(t, m, "Foo")
	assert.Equal(t, m["Foo"], "Bar")
	m = New(user).AddFuncIf(func(u interface{}) bool {
		return false
	}, "Foo", func(u interface{}) interface{} {
		return "Bar"
	}).Result()
	assert.NotContains(t, m, "Foo")
}

func TestConvert(t *testing.T) {
	m := New(user).Convert("FirstName", func(s interface{}) interface{} {
		return strings.ToLower(s.(string))
	}).Result()
	assert.Contains(t, m, "FirstName")
	assert.Equal(t, m["FirstName"], "foo")
}

func TestConvertIf(t *testing.T) {
	m := New(user).ConvertIf(func(u interface{}) bool {
		return true
	}, "FirstName", func(s interface{}) interface{} {
		return strings.ToLower(s.(string))
	}).Result()
	assert.Contains(t, m, "FirstName")
	assert.Equal(t, m["FirstName"], "foo")
	m = New(user).ConvertIf(func(u interface{}) bool {
		return false
	}, "FirstName", func(s interface{}) interface{} {
		return strings.ToLower(s.(string))
	}).Result()
	assert.NotContains(t, m, "FirstName")
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

func TestCustomSerializer(t *testing.T) {
	m := NewCustomSerializer(user).WithPrivateinfo().WithBasicInfo().Result()
	for _, field := range []string{"ID", "FirstName", "LastName", "Email"} {
		assert.Contains(t, m, field)
	}
}
