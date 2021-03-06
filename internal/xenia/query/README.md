

# query
`import "github.com/coralproject/shelf/internal/xenia/query"`

* [Overview](#pkg-overview)
* [Index](#pkg-index)
* [Subdirectories](#pkg-subdirectories)

## <a name="pkg-overview">Overview</a>
Package query provides the service layer for building apps using
query functionality.




## <a name="pkg-index">Index</a>
* [Constants](#pkg-constants)
* [Variables](#pkg-variables)
* [func Delete(context interface{}, db *db.DB, name string) error](#Delete)
* [func EnsureIndexes(context interface{}, db *db.DB, set *Set) error](#EnsureIndexes)
* [func GetAll(context interface{}, db *db.DB, tags []string) ([]Set, error)](#GetAll)
* [func GetNames(context interface{}, db *db.DB) ([]string, error)](#GetNames)
* [func Upsert(context interface{}, db *db.DB, set *Set) error](#Upsert)
* [type Index](#Index)
* [type Param](#Param)
* [type Query](#Query)
  * [func (q *Query) Validate() error](#Query.Validate)
* [type Result](#Result)
* [type Set](#Set)
  * [func GetByName(context interface{}, db *db.DB, name string) (*Set, error)](#GetByName)
  * [func GetLastHistoryByName(context interface{}, db *db.DB, name string) (*Set, error)](#GetLastHistoryByName)
  * [func (s *Set) PrepareForInsert()](#Set.PrepareForInsert)
  * [func (s *Set) PrepareForUse()](#Set.PrepareForUse)
  * [func (s *Set) Validate() error](#Set.Validate)


#### <a name="pkg-files">Package files</a>
[model.go](/src/github.com/coralproject/shelf/internal/xenia/query/model.go) [prepare.go](/src/github.com/coralproject/shelf/internal/xenia/query/prepare.go) [query.go](/src/github.com/coralproject/shelf/internal/xenia/query/query.go) 


## <a name="pkg-constants">Constants</a>
``` go
const (
    Collection        = "query_sets"
    CollectionHistory = "query_sets_history"
)
```
Contains the name of Mongo collections.

``` go
const (
    TypePipeline = "pipeline"
)
```
Set of query types we expect to receive.


## <a name="pkg-variables">Variables</a>
``` go
var (
    ErrNotFound = errors.New("Set Not found")
)
```
Set of error variables.



## <a name="Delete">func</a> [Delete](/src/target/query.go?s=9238:9300#L348)
``` go
func Delete(context interface{}, db *db.DB, name string) error
```
Delete is used to remove an existing Set document.



## <a name="EnsureIndexes">func</a> [EnsureIndexes](/src/target/query.go?s=1196:1262#L38)
``` go
func EnsureIndexes(context interface{}, db *db.DB, set *Set) error
```
EnsureIndexes perform index create commands against Mongo for the indexes
specied in each query for the set. It will attempt to ensure all indexes
regardless if one fails. Then reports all failures.



## <a name="GetAll">func</a> [GetAll](/src/target/query.go?s=5707:5780#L218)
``` go
func GetAll(context interface{}, db *db.DB, tags []string) ([]Set, error)
```
GetAll retrieves a list of sets.



## <a name="GetNames">func</a> [GetNames](/src/target/query.go?s=4579:4642#L172)
``` go
func GetNames(context interface{}, db *db.DB) ([]string, error)
```
GetNames retrieves a list of query names.



## <a name="Upsert">func</a> [Upsert](/src/target/query.go?s=2412:2471#L85)
``` go
func Upsert(context interface{}, db *db.DB, set *Set) error
```
Upsert is used to create or update an existing Set document.




## <a name="Index">type</a> [Index](/src/target/model.go?s=866:1578#L25)
``` go
type Index struct {
    Key        []string `bson:"key" json:"key"`                                   // Index key fields; prefix name with dash (-) for descending order
    Unique     bool     `bson:"unique,omitempty" json:"unique,omitempty"`         // Prevent two documents from having the same index key
    DropDups   bool     `bson:"drop_dups,omitempty" json:"drop_dups,omitempty"`   // Drop documents with the same index key as a previously indexed one
    Background bool     `bson:"background,omitempty" json:"background,omitempty"` // Build index in background and return immediately
    Sparse     bool     `bson:"sparse,omitempty" json:"sparse,omitempty"`         // Only index documents containing the Key fields
}
```
Index contains metadata for creating indexes in Mongo.










## <a name="Param">type</a> [Param](/src/target/model.go?s=3825:4196#L72)
``` go
type Param struct {
    Name      string `bson:"name" json:"name"`             // Name of the parameter.
    Desc      string `bson:"desc" json:"desc"`             // Description about the parameter.
    Default   string `bson:"default" json:"default"`       // Default value for the parameter.
    RegexName string `bson:"regex_name" json:"regex_name"` // Regular expression name.
}
```
Param contains meta-data about a required parameter for the query.










## <a name="Query">type</a> [Query](/src/target/model.go?s=1719:3282#L36)
``` go
type Query struct {
    Name        string                   `bson:"name" json:"name" validate:"required,min=3"`                                 // Unique name per query document.
    Description string                   `bson:"desc,omitempty" json:"desc,omitempty"`                                       // Description of this specific query.
    Type        string                   `bson:"type" json:"type" validate:"required,min=8"`                                 // TypePipeline, TypeTemplate
    Collection  string                   `bson:"collection,omitempty" json:"collection,omitempty" validate:"required,min=3"` // Name of the collection to use for processing the query.
    Timeout     string                   `bson:"timeout,omitempty" json:"timeout,omitempty"`                                 // Provides a timeout for the query if it does not return.
    Commands    []map[string]interface{} `bson:"commands" json:"commands"`                                                   // Commands to process for the query.
    Indexes     []Index                  `bson:"indexes" json:"indexes"`                                                     // Set of indexes required to optimize the execution of the query.
    Continue    bool                     `bson:"continue,omitempty" json:"continue,omitempty"`                               // Indicates that on failure to process the next query.
    Return      bool                     `bson:"return" json:"return"`                                                       // Return the results back to the user with Name as the key.
}
```
Query contains the configuration details for a query.










### <a name="Query.Validate">func</a> (\*Query) [Validate](/src/target/model.go?s=3336:3368#L49)
``` go
func (q *Query) Validate() error
```
Validate checks the query value for consistency.




## <a name="Result">type</a> [Result](/src/target/model.go?s=664:724#L18)
``` go
type Result struct {
    Results interface{} `json:"results"`
}
```
Result contains the result of an query set execution.
This had more fields in the past that have been removed. We
can't change this out without breaking the API.










## <a name="Set">type</a> [Set](/src/target/model.go?s=4338:5209#L82)
``` go
type Set struct {
    Name        string  `bson:"name" json:"name" validate:"required,min=3"` // Name of the query set.
    Description string  `bson:"desc" json:"desc"`                           // Description of the query set.
    PreScript   string  `bson:"pre_script" json:"pre_script"`               // Name of a script document to prepend.
    PstScript   string  `bson:"pst_script" json:"pst_script"`               // Name of a script document to append.
    Params      []Param `bson:"params" json:"params"`                       // Collection of parameters.
    Queries     []Query `bson:"queries" json:"queries"`                     // Collection of queries.
    Enabled     bool    `bson:"enabled" json:"enabled"`                     // If the query set is enabled to run.
    Explain     bool    `bson:"explain" json:"explain"`                     // If we want the explain output.
}
```
Set contains the configuration details for a rule set.







### <a name="GetByName">func</a> [GetByName](/src/target/query.go?s=6792:6865#L260)
``` go
func GetByName(context interface{}, db *db.DB, name string) (*Set, error)
```
GetByName retrieves the document for the specified Set.


### <a name="GetLastHistoryByName">func</a> [GetLastHistoryByName](/src/target/query.go?s=7755:7839#L296)
``` go
func GetLastHistoryByName(context interface{}, db *db.DB, name string) (*Set, error)
```
GetLastHistoryByName gets the last written Set within the history.





### <a name="Set.PrepareForInsert">func</a> (\*Set) [PrepareForInsert](/src/target/model.go?s=5520:5552#L109)
``` go
func (s *Set) PrepareForInsert()
```
PrepareForInsert replaces the documents for insertion.




### <a name="Set.PrepareForUse">func</a> (\*Set) [PrepareForUse](/src/target/model.go?s=5793:5822#L120)
``` go
func (s *Set) PrepareForUse()
```
PrepareForUse replaces the documents back to their orginal form.




### <a name="Set.Validate">func</a> (\*Set) [Validate](/src/target/model.go?s=5261:5291#L94)
``` go
func (s *Set) Validate() error
```
Validate checks the set value for consistency.








- - -
Generated by [godoc2md](http://godoc.org/github.com/davecheney/godoc2md)
