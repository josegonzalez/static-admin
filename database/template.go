package database

import (
	"time"

	"gorm.io/gorm"
)

type Template struct {
	gorm.Model
	UserID uint   `gorm:"not null;index:idx_user_template,priority:1"`
	Name   string `gorm:"not null"`
}

type TemplateField struct {
	gorm.Model
	TemplateID       uint             `gorm:"not null;index:idx_template_field,priority:1"`
	Name             string           `gorm:"not null"`
	StringValue      string           `gorm:"not null;default:''"`
	BoolValue        bool             `gorm:"not null;default:false"`
	NumberValue      float64          `gorm:"not null;default:0"`
	DateTimeValue    time.Time        `gorm:"not null;default:0000-00-00 00:00:00"`
	StringSliceValue StringSliceValue `gorm:"not null;default:'[]';serializer:json"`
	Type             string           `gorm:"not null;default:'string';check:type IN ('string', 'bool', 'number', 'dateTime', 'stringSlice')"`
}

type StringSliceValue []string
