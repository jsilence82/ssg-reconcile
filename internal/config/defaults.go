package config

const (
	DefaultCategoryGeneral = "Admission - General Admission"
	DefaultCategoryStudent = "Admission - Student"
	DefaultCategoryComp    = "Admission - Comp"
)

var DefaultPayPalPIIColumns = []string{
	"Name",
	"Address 1",
	"Address 2",
	"City",
	"State",
	"Zip",
	"Country",
	"Contact Phone Number",
	"Subject",
	"Note",
}

var DefaultTTPIIColumns = []string{
	"Attendee name",
	"Postcode / Zip",
}
