package image

type Dimg struct {
	*Stage
}

func NewDimgImage(from *Stage, name string, builtId string) *Dimg {
	dimg := &Dimg{}
	dimg.Stage = NewStageImage(from, name, builtId)
	return dimg
}

func (i *Dimg) Tag() error {
	return Tag(i.BuiltId, i.Name)
}

func (i *Dimg) Export() error {
	if err := i.Push(); err != nil {
		return err
	}

	if err := i.Untag(); err != nil {
		return err
	}

	return nil
}
