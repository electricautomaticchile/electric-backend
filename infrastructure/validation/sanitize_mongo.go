package validation

import (
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	mongoOperatorRegex = regexp.MustCompile(`^\$`)
)

func SanitizeMongoQuery(query bson.M) bson.M {
	sanitized := bson.M{}
	
	for key, value := range query {
		if mongoOperatorRegex.MatchString(key) {
			continue
		}
		
		sanitizedKey := strings.TrimSpace(key)
		
		switch v := value.(type) {
		case string:
			sanitized[sanitizedKey] = SanitizeString(v)
		case bson.M:
			sanitized[sanitizedKey] = SanitizeMongoQuery(v)
		case map[string]interface{}:
			sanitized[sanitizedKey] = SanitizeMongoQuery(bson.M(v))
		case []interface{}:
			sanitizedArray := make([]interface{}, len(v))
			for i, item := range v {
				if itemMap, ok := item.(bson.M); ok {
					sanitizedArray[i] = SanitizeMongoQuery(itemMap)
				} else if itemStr, ok := item.(string); ok {
					sanitizedArray[i] = SanitizeString(itemStr)
				} else {
					sanitizedArray[i] = item
				}
			}
			sanitized[sanitizedKey] = sanitizedArray
		default:
			sanitized[sanitizedKey] = value
		}
	}
	
	return sanitized
}

func ValidateObjectID(id string) (primitive.ObjectID, error) {
	id = strings.TrimSpace(id)
	return primitive.ObjectIDFromHex(id)
}

func SanitizeMongoFilter(filter interface{}) interface{} {
	switch f := filter.(type) {
	case bson.M:
		return SanitizeMongoQuery(f)
	case map[string]interface{}:
		return SanitizeMongoQuery(bson.M(f))
	default:
		return filter
	}
}
