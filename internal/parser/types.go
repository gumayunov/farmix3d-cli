package parser

type Transform3D struct {
	Matrix [12]float64 `json:"matrix"`
}

type ComponentInfo struct {
	ID         int         `json:"id"`
	Name       string      `json:"name"`
	SourceFile string      `json:"source_file"`
	Transform  Transform3D `json:"transform"`
}

type PlateObject struct {
	ID         int             `json:"id"`
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	Material   string          `json:"material,omitempty"`
	Position   Transform3D     `json:"position"`
	Printable  bool            `json:"printable"`
	Components []ComponentInfo `json:"components,omitempty"`
}

type PlateInfo struct {
	PlateID   int           `json:"plate_id"`
	PlateName string        `json:"plate_name"`
	Objects   []PlateObject `json:"objects"`
}

type GroupedObject struct {
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	Material   string          `json:"material,omitempty"`
	Count      int             `json:"count"`
	Components []ComponentInfo `json:"components,omitempty"`
	ObjectIDs  []int           `json:"object_ids"`
}

type Parser3MF struct {
	Plates []PlateInfo `json:"plates"`
}

type ModelObject struct {
	ID         int                    `xml:"id,attr"`
	Type       string                 `xml:"type,attr"`
	Name       string                 `xml:"name,attr"`
	Mesh       *Mesh                  `xml:"mesh"`
	Components *ComponentsCollection  `xml:"components"`
}

type Mesh struct {
	Vertices  []Vertex   `xml:"vertices>vertex"`
	Triangles []Triangle `xml:"triangles>triangle"`
}

type Vertex struct {
	X float64 `xml:"x,attr"`
	Y float64 `xml:"y,attr"`
	Z float64 `xml:"z,attr"`
}

type Triangle struct {
	V1 int `xml:"v1,attr"`
	V2 int `xml:"v2,attr"`
	V3 int `xml:"v3,attr"`
}

type ComponentsCollection struct {
	Components []Component `xml:"component"`
}

type Component struct {
	ObjectID  int    `xml:"objectid,attr"`
	Transform string `xml:"transform,attr"`
	Path      string `xml:"path,attr"`
}

type BuildItem struct {
	ObjectID  int    `xml:"objectid,attr"`
	Transform string `xml:"transform,attr"`
	Printable *bool  `xml:"printable,attr"`
}

type Model3D struct {
	Unit      string        `xml:"unit,attr"`
	Resources []ModelObject `xml:"resources>object"`
	Build     []BuildItem   `xml:"build>item"`
}

type Plate struct {
	PlaterID   int             `xml:"plater_id,attr"`
	PlaterName string          `xml:"plater_name,attr"`
	Metadata   []MetadataEntry `xml:"metadata"`
	Instances  []ModelInstance `xml:"model_instance"`
}

type ModelInstance struct {
	ObjectID   int             `xml:"object_id,attr"`
	InstanceID int             `xml:"instance_id,attr"`
	IdentifyID int             `xml:"identify_id,attr"`
	Metadata   []MetadataEntry `xml:"metadata"`
}

type MetadataEntry struct {
	Key   string `xml:"key,attr"`
	Value string `xml:"value,attr"`
}

type ObjectMeta struct {
	ID       int             `xml:"id,attr"`
	Name     string          `xml:"name,attr"`
	Metadata []MetadataEntry `xml:"metadata"`
	Parts    []PartMeta      `xml:"part"`
}

type PartMeta struct {
	ID         int             `xml:"id,attr"`
	Name       string          `xml:"name,attr"`
	SourceFile string          `xml:"source_file,attr"`
	Metadata   []MetadataEntry `xml:"metadata"`
}

type FilamentSettings struct {
	Name string `json:"name"`
}

type ModelSettings struct {
	Plates        []Plate         `xml:"plate"`
	Instances     []ModelInstance `xml:"model_instance"`
	Objects       []ObjectMeta    `xml:"object"`
	Parts         []PartMeta      `xml:"part"`
}