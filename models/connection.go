package irminmodels

type CustomFieldValues map[string]string

type Connection struct {
	ID            string            `json:"id"             validate:"required,validsqid=connections" example:"conn_5p8q2n7m9x4k"`
	Name          string            `json:"name"           validate:"required,max=100"               example:"Production MySQL Database"`
	Description   string            `json:"description"    validate:"max=500"                        example:"Primary MySQL database for production customer data"`
	Documentation string            `json:"documentation"  validate:"validdocumentation"             example:"# Production Database\n\nThis connection provides access to the main customer database..."`
	Details       CustomFieldValues `json:"details"        validate:"required"                       example:"{\"host\":\"db.example.com\",\"port\":\"3306\",\"database\":\"customers\"}"`
	Settings      CustomFieldValues `json:"settings"       validate:"required"                       example:"{\"ssl_enabled\":\"true\",\"charset\":\"utf8mb4\",\"timeout\":\"30\"}"`
	Owner         User              `json:"owner"          validate:"required"`
	Connector     Connector         `json:"connector"      validate:"required"`
	Tags          []Tag             `json:"tags,omitempty" validate:"dive"`
}
