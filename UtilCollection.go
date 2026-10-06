package utils

type utilCollection struct {
}

func (_self *utilCollection) IntArrayFind(slice []int, value int) int {
	for p, v := range slice {
		if v == value {
			return p
		}
	}
	return -1
}

func (_self *utilCollection) IntArrayContain(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func (_self *utilCollection) Int64ArrayFind(slice []int64, value int64) int {
	for p, v := range slice {
		if v == value {
			return p
		}
	}
	return -1
}

func (_self *utilCollection) Int64ArrayContain(slice []int64, value int64) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func (_self *utilCollection) StrArrayContain(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
