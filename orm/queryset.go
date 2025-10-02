package orm

import (
	"gorm.io/gorm"
)

// QuerySet provides Django-like queryset API on top of GORM
type QuerySet struct {
	db *gorm.DB
}

// NewQuerySet creates a new QuerySet
func NewQuerySet(db *gorm.DB) *QuerySet {
	return &QuerySet{db: db}
}

// All returns all records
func (q *QuerySet) All(dest interface{}) error {
	return q.db.Find(dest).Error
}

// Filter filters records by conditions
func (q *QuerySet) Filter(query interface{}, args ...interface{}) *QuerySet {
	return &QuerySet{db: q.db.Where(query, args...)}
}

// Exclude excludes records by conditions
func (q *QuerySet) Exclude(query interface{}, args ...interface{}) *QuerySet {
	return &QuerySet{db: q.db.Not(query, args...)}
}

// Get retrieves a single record
func (q *QuerySet) Get(dest interface{}) error {
	return q.db.First(dest).Error
}

// Count returns the count of records
func (q *QuerySet) Count() (int64, error) {
	var count int64
	err := q.db.Count(&count).Error
	return count, err
}

// Exists checks if records exist
func (q *QuerySet) Exists() (bool, error) {
	count, err := q.Count()
	return count > 0, err
}

// OrderBy orders the results
func (q *QuerySet) OrderBy(fields ...string) *QuerySet {
	db := q.db
	for _, field := range fields {
		db = db.Order(field)
	}
	return &QuerySet{db: db}
}

// Limit limits the number of results
func (q *QuerySet) Limit(limit int) *QuerySet {
	return &QuerySet{db: q.db.Limit(limit)}
}

// Offset sets the offset for results
func (q *QuerySet) Offset(offset int) *QuerySet {
	return &QuerySet{db: q.db.Offset(offset)}
}

// Select specifies fields to retrieve
func (q *QuerySet) Select(fields ...string) *QuerySet {
	return &QuerySet{db: q.db.Select(fields)}
}

// Preload preloads associations
func (q *QuerySet) Preload(associations ...string) *QuerySet {
	db := q.db
	for _, assoc := range associations {
		db = db.Preload(assoc)
	}
	return &QuerySet{db: db}
}

// Create creates a new record
func (q *QuerySet) Create(value interface{}) error {
	return q.db.Create(value).Error
}

// Update updates records
func (q *QuerySet) Update(column string, value interface{}) error {
	return q.db.Update(column, value).Error
}

// Updates updates multiple columns
func (q *QuerySet) Updates(values interface{}) error {
	return q.db.Updates(values).Error
}

// Delete deletes records
func (q *QuerySet) Delete(value interface{}) error {
	return q.db.Delete(value).Error
}

// First gets the first record
func (q *QuerySet) First(dest interface{}) error {
	return q.db.First(dest).Error
}

// Last gets the last record
func (q *QuerySet) Last(dest interface{}) error {
	return q.db.Last(dest).Error
}