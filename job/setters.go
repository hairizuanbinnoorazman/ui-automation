package job

func SetStatus(status Status) UpdateSetter {
	return func(j *Job) error {
		if !status.IsValid() {
			return ErrInvalidStatus
		}
		j.Status = status
		return nil
	}
}

func SetConfig(config JSONMap) UpdateSetter {
	return func(j *Job) error {
		j.Config = config
		return nil
	}
}

func SetResult(result JSONMap) UpdateSetter {
	return func(j *Job) error {
		j.Result = result
		return nil
	}
}
