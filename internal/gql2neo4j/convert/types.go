package convert

// Neo4jEdition represents the edition of the Neo4j database.
type Neo4jEdition uint8

const (
	// Neo4jEnterpriseEdition represents the enterprise edition of Neo4j.
	Neo4jEnterpriseEdition Neo4jEdition = iota
	// Neo4jCommunityEdition represents the community edition of Neo4j.
	Neo4jCommunityEdition
)

// EntityType represents the type of entity in Neo4j.
type EntityType uint8

const (
	// NodeEntityType represents a node entity type.
	NodeEntityType EntityType = iota
	// EdgeEntityType represents an edge entity type.
	EdgeEntityType
)

// DatabaseMetadata holds metadata about the Neo4j database.
type DatabaseMetadata struct {
	Edition     Neo4jEdition
	Version     string
	APOCEnabled bool
}
