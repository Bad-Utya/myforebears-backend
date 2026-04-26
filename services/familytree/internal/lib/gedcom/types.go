package gedcom

// GEDCOMPerson represents a person in GEDCOM format
type GEDCOMPerson struct {
	ID            string
	FirstName     string
	LastName      string
	Patronymic    string
	Gender        string   // M, F, or U
	FamilySpouses []string // IDs of spouses (FAMS)
	FamilyAsChild string   // ID of family as child (FAMC)
}

// GEDCOMFamily represents a family relationship in GEDCOM format
type GEDCOMFamily struct {
	ID       string
	Husband  string   // ID of husband
	Wife     string   // ID of wife
	Children []string // IDs of children
}

// GEDCOMData represents the complete tree data in GEDCOM-ready format
type GEDCOMData struct {
	Persons      map[string]*GEDCOMPerson // ID -> Person
	Families     map[string]*GEDCOMFamily // ID -> Family
	NextPersonID int
	NextFamilyID int
}
