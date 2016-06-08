
# cmddb
    import "github.com/coralproject/xenia/cmd/xenia/cmddb"





## Variables
``` go
var (
    // ErrCollectionExists is return when a collection to be
    // created already exists.
    ErrCollectionExists = errors.New("Collection already exists.")
)
```

## func GetCommands
``` go
func GetCommands(db *db.DB) *cobra.Command
```
GetCommands returns the db commands.



## type Collection
``` go
type Collection struct {
    Name    string  `json:"name"`
    Indexes []Index `json:"indexes"`
}
```
Collection is the container for a db collection definition.











## type DBMeta
``` go
type DBMeta struct {
    Cols []Collection `json:"collections"`
}
```
DBMeta is the container for all db objects.











## type Field
``` go
type Field struct {
    Name      string `json:"name"`
    Type      int    `json:"type"`
    OtherType string `json:"other"`
}
```
Field is the container for a field definition.











## type Index
``` go
type Index struct {
    Name     string  `json:"name"`
    IsUnique bool    `json:"unique"`
    Fields   []Field `json:"fields"`
}
```
Index is the container for an index definition.

















- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)