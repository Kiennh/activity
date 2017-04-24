package activity

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/qor/test/utils"
	"github.com/theplant/testingutils"
)

var db *gorm.DB
var Admin *admin.Admin

var (
	mux    = http.NewServeMux()
	Server = httptest.NewServer(mux)
)

type Order struct {
	gorm.Model
	Code string
}

func init() {
	db = utils.TestDB()
	if err := db.DropTableIfExists(&QorActivity{}, &Order{}).Error; err != nil {
		panic(err)
	}
	db.AutoMigrate(&QorActivity{}, &Order{})
	db.Create(&Order{Code: "1000001"})
	Admin = admin.New(&qor.Config{DB: db})
	orderRes := Admin.AddResource(&Order{})
	Register(orderRes)
	Admin.MountTo("/admin", mux)
}

func TestCRUDActivity(t *testing.T) {
	http.PostForm(Server.URL+"/admin/orders/1/!qor_activities", url.Values{
		"QorResource.Action":  {"comment on"},
		"QorResource.Content": {"Activity Title"},
		"QorResource.Note":    {"Activity Note"},
	})

	req, _ := http.Get(Server.URL + "/admin/orders/1/!qor_activities.json")
	assertActivityEqual(t, req.Body, QorActivity{
		Action:       "comment on",
		Content:      "Activity Title",
		Note:         "Activity Note",
		ResourceID:   "1",
		ResourceType: "orders",
	})

	http.PostForm(Server.URL+"/admin/orders/1/!qor_activities/1/edit", url.Values{
		"QorResource.ID":   {"1"},
		"QorResource.Note": {"Activity Note Changed"},
	})

	req, _ = http.Get(Server.URL + "/admin/orders/1/!qor_activities.json")
	assertActivityEqual(t, req.Body, QorActivity{
		Action:       "comment on",
		Content:      "Activity Title",
		Note:         "Activity Note Changed",
		ResourceID:   "1",
		ResourceType: "orders",
	})
}

func assertActivityEqual(t *testing.T, results io.ReadCloser, act QorActivity) {
	records := []*QorActivity{}
	content, _ := ioutil.ReadAll(results)
	json.Unmarshal(content, &records)
	for _, record := range records {
		record.ID = 0
	}
	diff := testingutils.PrettyJsonDiff(records, []QorActivity{act})
	if len(diff) > 0 {
		t.Errorf(diff)
	}
}
