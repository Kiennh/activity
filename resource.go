package activity

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
	"github.com/qor/qor"
)

type SetResourceAble interface {
	SetResourceType(string)
	SetResourceID(string)
	SetCreated(qor.CurrentUser)
}

func prepareGetActivitiesDB(context *admin.Context, result interface{}, types ...string) *gorm.DB {
	var baseStruct = context.Resource.NewStruct()
	resourceID := getPrimaryKey(context, result)
	db := context.GetDB().Model(baseStruct).Order("id asc").Where("resource_id = ? AND resource_type = ?", resourceID, context.Resource.ToParam())

	var inTypes, notInTypes []string
	for _, t := range types {
		if strings.HasPrefix(t, "-") {
			notInTypes = append(notInTypes, strings.TrimPrefix(t, "-"))
		} else {
			inTypes = append(inTypes, t)
		}
	}

	if len(inTypes) > 0 {
		db = db.Where("type IN (?)", inTypes)
	}

	if len(notInTypes) > 0 {
		db = db.Where("type NOT IN (?)", notInTypes)
	}

	return db
}

// GetActivities get activities for selected types
func GetActivities(context *admin.Context, result interface{}, types ...string) (interface{}, error) {
	var activityResource = context.Admin.GetResource("QorActivity")
	var activities = activityResource.NewSlice()
	db := prepareGetActivitiesDB(context, result, types...)
	err := db.Find(activities).Error
	return activities, err
}

// GetActivitiesCount get activities's count for selected types
func GetActivitiesCount(context *admin.Context, result interface{}, types ...string) int {
	var count int
	prepareGetActivitiesDB(context, result, types...).Model(&QorActivity{}).Count(&count)
	return count
}

// CreateActivity creates an activity for this context
func CreateActivity(context *admin.Context, activity interface{}, result interface{}) error {
	var activityResource = context.Admin.GetResource("QorActivity")

	// fill in necessary activity fields
	if setter, ok := activity.(SetResourceAble); ok {
		setter.SetResourceType(context.Resource.ToParam())
		setter.SetResourceID(getPrimaryKey(context, result))
		if context.CurrentUser != nil {
			setter.SetCreated(context.CurrentUser)
		}

	}

	return activityResource.CallSave(activity, context.Context)
}

func getPrimaryKey(context *admin.Context, record interface{}) string {
	db := context.GetDB()

	var primaryValues []string
	for _, field := range db.NewScope(record).PrimaryFields() {
		primaryValues = append(primaryValues, fmt.Sprint(field.Field.Interface()))
	}
	return strings.Join(primaryValues, "::")
}
