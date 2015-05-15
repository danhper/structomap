package serializer

import (
	"encoding/json"
	"fmt"
	"github.com/tuvistavie/testify/assert"
	"testing"
	"time"
)

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

func NewCustomSerializer() *CustomSerializer {
	return &CustomSerializer{New()}
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
	userSerializer := New().
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
	})
	userMap := userSerializer.Transform(user)
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
	m := New().PickAll().Transform(user)
	assert.Contains(t, m, "ID")
	assert.Contains(t, m, "Age")
	assert.Contains(t, m, "FirstName")
}

func TestPick(t *testing.T) {
	m := New().Pick("ID", "Age").Transform(user)
	assert.Contains(t, m, "ID")
	assert.Contains(t, m, "Age")
	assert.NotContains(t, m, "FirstName")
	m = New().Pick("ID").Pick("Age").Transform(user)
	assert.Contains(t, m, "ID")
	assert.Contains(t, m, "Age")
	assert.Equal(t, m["ID"], user.ID)
}

func TestPickIf(t *testing.T) {
	m := New().
		PickIf(alwaysTrue, "ID", "FirstName").
		PickIf(alwaysFalse, "Email").Transform(user)
	assert.Contains(t, m, "ID")
	assert.Contains(t, m, "FirstName")
	assert.NotContains(t, m, "Email")
	assert.NotContains(t, m, "Age")
}

func TestPickFunc(t *testing.T) {
	m := New().PickFunc(func(t interface{}) interface{} {
		return t.(time.Time).Format(time.RFC3339)
	}, "CreatedAt", "UpdatedAt").Transform(user)
	assert.Contains(t, m, "CreatedAt")
	assert.Contains(t, m, "UpdatedAt")
	assert.Equal(t, m["CreatedAt"], user.CreatedAt.Format(time.RFC3339))
}

func TestPickFuncIf(t *testing.T) {
	m := New().PickFuncIf(alwaysTrue, func(t interface{}) interface{} {
		return t.(time.Time).Format(time.RFC3339)
	}, "CreatedAt", "UpdatedAt").PickFuncIf(alwaysFalse, identity, "Email").Transform(user)
	assert.Contains(t, m, "CreatedAt")
	assert.Contains(t, m, "UpdatedAt")
	assert.NotContains(t, m, "Email")
	assert.Equal(t, m["CreatedAt"], user.CreatedAt.Format(time.RFC3339))
}

func TestOmit(t *testing.T) {
	m := New().PickAll().Omit("Birthday", "FirstName").Transform(user)
	assert.NotContains(t, m, "Birthday")
	assert.NotContains(t, m, "FirstName")
	assert.Contains(t, m, "ID")
}

func TestOmitIf(t *testing.T) {
	m := New().PickAll().OmitIf(func(u interface{}) bool {
		return u.(User).HideName
	}, "FirstName", "LastName").Transform(user)
	assert.NotContains(t, m, "FirstName", "LastName")

	m = New().PickAll().OmitIf(func(u interface{}) bool {
		return u.(User).Age < 18
	}, "Birthday", "Age").Transform(user)
	assert.Contains(t, m, "Birthday")
	assert.Contains(t, m, "Age")
}

func TestAdd(t *testing.T) {
	m := New().Add("Foo", "Bar").Transform(user)
	assert.Contains(t, m, "Foo")
	assert.Equal(t, "Bar", m["Foo"])
}

func TestAddIf(t *testing.T) {
	m := New().AddIf(alwaysFalse, "Foo", "Bar").Transform(user)
	assert.NotContains(t, m, "Foo")
	m = New().AddIf(alwaysTrue, "Foo", "Bar").Transform(user)
	assert.Contains(t, m, "Foo")
	assert.Equal(t, "Bar", m["Foo"])
}

func TestAddFunc(t *testing.T) {
	m := New().AddFunc("FullName", func(u interface{}) interface{} {
		return u.(User).FirstName + " " + u.(User).LastName
	}).Transform(user)
	assert.Contains(t, m, "FullName")
	assert.Equal(t, user.FirstName+" "+user.LastName, m["FullName"])
}

func TestAddFuncIf(t *testing.T) {
	m := New().AddFuncIf(alwaysTrue, "Foo", func(u interface{}) interface{} {
		return "Bar"
	}).Transform(user)
	assert.Contains(t, m, "Foo")
	assert.Equal(t, m["Foo"], "Bar")
	m = New().AddFuncIf(alwaysFalse, "Foo", identity).Transform(user)
	assert.NotContains(t, m, "Foo")
}

func TestConvertKeys(t *testing.T) {
	m := New().PickAll().ConvertKeys(func(s string) string {
		return "dummy_" + s
	}).Transform(user)
	assert.Contains(t, m, "dummy_ID")
	assert.Contains(t, m, "dummy_FirstName")
	assert.NotContains(t, m, "ID")
}

func TestUseSnakeCase(t *testing.T) {
	m := New().UseSnakeCase().PickAll().Transform(user)
	assert.Contains(t, m, "id")
	assert.Contains(t, m, "first_name")
	assert.NotContains(t, m, "FirstName")
}

func TestUseCamelCase(t *testing.T) {
	m := New().UseCamelCase().PickAll().Transform(user)
	assert.Contains(t, m, "id")
	assert.Contains(t, m, "firstName")
	assert.NotContains(t, m, "FirstName")
}

func TestUsePascalCase(t *testing.T) {
	m := New().UsePascalCase().PickAll().Transform(user)
	assert.Contains(t, m, "ID")
	assert.Contains(t, m, "FirstName")
}

func TestDefaultCase(t *testing.T) {
	SetDefaultCase(SnakeCase)
	m := New().PickAll().Transform(user)
	assert.Contains(t, m, "first_name")
	SetDefaultCase(CamelCase)
	m = New().PickAll().Transform(user)
	assert.Contains(t, m, "firstName")
	SetDefaultCase(PascalCase)
	m = New().PickAll().Transform(user)
	assert.Contains(t, m, "FirstName")
}

func TestTranformArray(t *testing.T) {
	otherUser := User{ID: 8, FirstName: "Me"}
	users := []interface{}{user, otherUser}
	arr := New().Pick("ID", "FirstName").TransformArray(users)
	assert.Len(t, arr, 2)
	assert.Equal(t, arr[0]["ID"], user.ID)
	assert.Equal(t, arr[1]["ID"], otherUser.ID)
}

func TestCustomSerializer(t *testing.T) {
	m := NewCustomSerializer().WithPrivateinfo().WithBasicInfo().Transform(user)
	for _, field := range []string{"ID", "FirstName", "LastName", "Email"} {
		assert.Contains(t, m, field)
	}
}
