package fieldbinding

import (
	"database/sql"
	"sync"

	mssqldriver "github.com/denisenkom/go-mssqldb"
)

// NewFieldBinding initializes the struct
func NewFieldBinding() *FieldBinding {
	return &FieldBinding{}
}

// FieldBinding is deisgned for SQL rows.Scan() query.
type FieldBinding struct {
	sync.RWMutex // embedded.  see http://golang.org/ref/spec#Struct_types
	FieldArr     []interface{}
	FieldPtrArr  []interface{}
	FieldTypeArr []*sql.ColumnType
	FieldCount   int64
	MapFieldToID map[string]int64
}

func (fb *FieldBinding) put(k string, v int64) {
	fb.Lock()
	defer fb.Unlock()
	fb.MapFieldToID[k] = v
}

// Get gets a field
func (fb *FieldBinding) Get(k string) interface{} {
	fb.RLock()
	defer fb.RUnlock()
	// TODO: check map key exist and fb.FieldArr boundary.
	return fb.FieldArr[fb.MapFieldToID[k]]
}

// PutFields puts fields into the arrays
func (fb *FieldBinding) PutFields(fArr []string) {
	fCount := len(fArr)
	fb.FieldArr = make([]interface{}, fCount)
	fb.FieldPtrArr = make([]interface{}, fCount)
	fb.MapFieldToID = make(map[string]int64, fCount)
	fb.FieldTypeArr = make([]*sql.ColumnType, fCount)

	for k, v := range fArr {
		fb.FieldPtrArr[k] = &fb.FieldArr[k]
		fb.put(v, int64(k))
	}
}

// GetFieldPtrArr returns the field pointer array
func (fb *FieldBinding) GetFieldPtrArr() []interface{} {
	return fb.FieldPtrArr
}

// GetFieldArr return the field array map
func (fb *FieldBinding) GetFieldArr() map[string]interface{} {
	m := make(map[string]interface{}, fb.FieldCount)

	for k, v := range fb.MapFieldToID {
		var u mssqldriver.UniqueIdentifier
		err := u.Scan(fb.FieldArr[v])
		if err != nil {
			m[k] = fb.FieldArr[v]
			continue
		}
		m[k] = u.String()
	}
	return m
}
