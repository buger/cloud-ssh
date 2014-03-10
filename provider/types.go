package provider

type Tag struct {
    Name, Value string
}

type Instances map[string][]Tag