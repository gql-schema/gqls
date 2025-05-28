// Package neo4j is a package that provides functionality to interact with Neo4j, a graph database.
// It includes structures and methods to create schema objects, constraints, and triggers in Neo4j.
package neo4j

// PropertyType represents the type of property in Neo4j.
type PropertyType string

//revive:disable:exported
const (
	Boolean       PropertyType = "BOOLEAN"
	Date          PropertyType = "DATE"
	Duration      PropertyType = "DURATION"
	Float         PropertyType = "FLOAT"
	Integer       PropertyType = "INTEGER"
	List          PropertyType = "LIST"
	LocalDateTime PropertyType = "LOCAL DATETIME"
	LocalTime     PropertyType = "LOCAL TIME"
	Point         PropertyType = "POINT"
	String        PropertyType = "STRING"
	ZonedDateTime PropertyType = "ZONED DATETIME"
	ZonedTime     PropertyType = "ZONED TIME"
)

//revive:enable:exported

// SchemaObject represents a schema object in Neo4j.
type SchemaObject interface {
	// CreateStatement generates the CREATE statement for the schema object.
	CreateStatement() string
}
