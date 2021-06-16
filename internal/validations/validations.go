package validations

type validator interface {
	Validate() error
}
func RunValidations(validations ...validator) error {
	for _, validation := range validations {
		if err := validation.Validate(); err != nil {
			return err
		}
	}
	return nil
}



