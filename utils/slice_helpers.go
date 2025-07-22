package utils

// UniqueUintSlice menghapus elemen duplikat dari sebuah slice uint.
func UniqueUintSlice(slice []uint) []uint {
	keys := make(map[uint]bool)
	var list []uint
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}