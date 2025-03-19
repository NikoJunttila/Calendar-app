package calendar

import (
	"gothstack/app/db"
	"gothstack/plugins/auth"
	"time"

	"gorm.io/gorm"
)

// WorkResource represents the work_resources table in the database
type WorkResource struct {
	ID                  uint           `gorm:"primaryKey"`
	Name                string         `gorm:"not null"`
	OwnerID             uint           `gorm:"not null"`
	CalendarID          uint           `gorm:"not null"`
	ResourcesPercentage int            `gorm:"not null"`
	CreatedAt           time.Time      `gorm:"not null"`
	UpdatedAt           time.Time      `gorm:"not null"`
	DeletedAt           gorm.DeletedAt `gorm:"index"`
	// Relationship fields
	Calendar Calendar  `gorm:"foreignKey:CalendarID"`
	Owner    auth.User `gorm:"foreignKey:OwnerID"`
}

// Event name constants
const (
	WorkResourceCreatedEvent = "work_resource.created"
	WorkResourceUpdatedEvent = "work_resource.updated"
	WorkResourceDeletedEvent = "work_resource.deleted"
)

// CreateWorkResource creates a new work resource
func CreateWorkResource(name string, ownerID uint, calendarID uint, resourcesPercentage int) (WorkResource, error) {
	resource := WorkResource{
		Name:                name,
		OwnerID:             ownerID,
		CalendarID:          calendarID,
		ResourcesPercentage: resourcesPercentage,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	result := db.Get().Create(&resource)
	return resource, result.Error
}

// GetWorkResource retrieves a work resource by its ID
func GetWorkResource(id uint) (WorkResource, error) {
	var resource WorkResource
	result := db.Get().First(&resource, id)
	return resource, result.Error
}

// GetWorkResourceWithRelations retrieves a work resource by ID along with its related calendar and owner
func GetWorkResourceWithRelations(id uint) (WorkResource, error) {
	var resource WorkResource
	result := db.Get().Preload("Calendar").Preload("Owner").First(&resource, id)
	return resource, result.Error
}

// ListWorkResources returns all work resources
func ListWorkResources() ([]WorkResource, error) {
	var resources []WorkResource
	result := db.Get().Find(&resources)
	return resources, result.Error
}

// ListWorkResourcesByOwner returns all work resources for a specific owner
func ListWorkResourcesByOwner(ownerID uint) ([]WorkResource, error) {
	var resources []WorkResource
	result := db.Get().Where("owner_id = ?", ownerID).Find(&resources)
	return resources, result.Error
}

// ListWorkResourcesByCalendar returns all work resources for a specific calendar
func ListWorkResourcesByCalendar(calendarID uint) ([]WorkResource, error) {
	var resources []WorkResource
	result := db.Get().Where("calendar_id = ?", calendarID).Find(&resources)
	return resources, result.Error
}

// UpdateWorkResource updates an existing work resource
func UpdateWorkResource(id uint, name string, resourcesPercentage int) (WorkResource, error) {
	var resource WorkResource
	if err := db.Get().First(&resource, id).Error; err != nil {
		return resource, err
	}

	resource.Name = name
	resource.ResourcesPercentage = resourcesPercentage
	resource.UpdatedAt = time.Now()

	result := db.Get().Save(&resource)
	return resource, result.Error
}

// DeleteWorkResource soft deletes a work resource by its ID
func DeleteWorkResource(id uint) error {
	result := db.Get().Delete(&WorkResource{}, id)
	return result.Error
}

// PermanentDeleteWorkResource permanently deletes a work resource by its ID
func PermanentDeleteWorkResource(id uint) error {
	result := db.Get().Unscoped().Delete(&WorkResource{}, id)
	return result.Error
}

// RestoreWorkResource restores a soft-deleted work resource
func RestoreWorkResource(id uint) (WorkResource, error) {
	var resource WorkResource
	result := db.Get().Unscoped().First(&resource, id)
	if result.Error != nil {
		return resource, result.Error
	}

	result = db.Get().Model(&resource).Update("deleted_at", nil)
	return resource, result.Error
}
