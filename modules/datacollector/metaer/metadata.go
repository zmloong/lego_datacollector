package metaer

type MetaDataBase struct {
	Name string
}

func (this *MetaDataBase) GetName() string {
	return this.Name
}
