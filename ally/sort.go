package ally

import (
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/TeamTutx/plib/perror"
)

var chunkifyRegexp = regexp.MustCompile(`(\d+|\D+)`)

func chunkify(s string) []string {
	return chunkifyRegexp.FindAllString(s, -1)
}

type modelDetails struct {
	Fields   []string
	ModSlice []interface{}
}

//NaturalSortCompare will compare returns true if the first string precedes the second one according to natural order
func NaturalSortCompare(a, b string) bool {
	chunksA := chunkify(a)
	chunksB := chunkify(b)

	nChunksA := len(chunksA)
	nChunksB := len(chunksB)

	for i := range chunksA {
		if i >= nChunksB {
			return false
		}

		aInt, aErr := strconv.Atoi(chunksA[i])
		bInt, bErr := strconv.Atoi(chunksB[i])

		// If both chunks are numeric, compare them as integers
		if aErr == nil && bErr == nil {
			if aInt == bInt {
				if i == nChunksA-1 {
					// We reached the last chunk of A, thus B is greater than A
					return true
				} else if i == nChunksB-1 {
					// We reached the last chunk of B, thus A is greater than B
					return false
				}

				continue
			}

			return aInt < bInt
		}

		// So far both strings are equal, continue to next chunk
		if chunksA[i] == chunksB[i] {
			if i == nChunksA-1 {
				// We reached the last chunk of A, thus B is greater than A
				return true
			} else if i == nChunksB-1 {
				// We reached the last chunk of B, thus A is greater than B
				return false
			}

			continue
		}

		return chunksA[i] < chunksB[i]
	}

	return false
}

func (md modelDetails) Len() int {
	return len(md.ModSlice)
}

func (md modelDetails) Less(a, b int) bool {
	return md.Compare(a, b, 0)
}

func (md modelDetails) Swap(a, b int) {
	sref := reflect.ValueOf(md.ModSlice[a])
	dref := reflect.ValueOf(md.ModSlice[b])
	temp := reflect.ValueOf(dref.Elem().Interface())
	dref.Elem().Set(reflect.ValueOf(sref.Elem().Interface()))
	sref.Elem().Set(temp)
}

// NatSort sorts a list of strings in a natural order
//Use `natsort:"1"` Tag with order of the feild as its value on model to use Natural sort for slice of structs
func NatSort(l interface{}) (err error) {
	var modDet modelDetails
	s := reflect.ValueOf(l)
	if s.Kind() == reflect.Slice && s.Len() > 0 {
		for i := 0; i < s.Len(); i++ {
			modDet.ModSlice = append(modDet.ModSlice, s.Index(i).Addr().Interface())
		}
		if reflect.ValueOf(modDet.ModSlice[0]).Elem().Kind() == reflect.Struct {
			if err = modDet.GetSortField(); err == nil {
				sort.Sort(modDet)
			}
		} else if reflect.ValueOf(modDet.ModSlice[0]).Elem().Kind() == reflect.String {
			sort.Sort(modDet)
		} else {
			err = perror.CustomError("Unsuported DataType For Sorting")
		}
	} else {
		err = perror.CustomError("Unsuported DataType For Sorting")
	}
	return
}

//GetSortField will get the names of the feilds to sort on
func (md *modelDetails) GetSortField() (err error) {
	t := reflect.TypeOf(md.ModSlice[0]).Elem()
	natsortOrderMap := make(map[int]string)
	order := make([]int, 0)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if valStr := field.Tag.Get("natsort"); valStr != "" {
			if field.Type.Kind() == reflect.String {
				if reflect.ValueOf(md.ModSlice[0]).Elem().FieldByName(field.Name).CanSet() {
					var val int
					if val, err = strconv.Atoi(valStr); err == nil {
						if _, exists := natsortOrderMap[val]; !exists {
							natsortOrderMap[val] = field.Name
							order = append(order, val)
						}
					} else {
						return
					}
				} else {
					err = perror.CustomError("All Sorting Feilds Must Be Exported")
					return
				}
			} else {
				err = perror.CustomError("All Sorting Feilds Must Be String")
				return
			}
		}
	}
	if len(order) > 0 {
		sort.Ints(order)
		for _, curOrder := range order {
			md.Fields = append(md.Fields, natsortOrderMap[curOrder])
		}
	} else {
		err = perror.CustomError("No Sorting Feilds Found")
	}
	return
}

// Compare returns true if the first string precedes the second one according to natural order
func (md *modelDetails) Compare(l, m, f int) bool {
	var a, b string
	ref1 := reflect.ValueOf(md.ModSlice[l]).Elem()
	ref2 := reflect.ValueOf(md.ModSlice[m]).Elem()

	if ref1.Kind() == reflect.Struct && ref2.Kind() == reflect.Struct {
		a, _ = ref1.FieldByName(md.Fields[f]).Interface().(string)
		b, _ = ref2.FieldByName(md.Fields[f]).Interface().(string)
		if a == b && f+1 < len(md.Fields) {
			return md.Compare(l, m, f+1)
		}
	} else if ref1.Kind() == reflect.String && ref2.Kind() == reflect.String {
		a = ref1.Interface().(string)
		b = ref2.Interface().(string)
	}

	chunksA := chunkify(strings.ToLower(a))
	chunksB := chunkify(strings.ToLower(b))

	nChunksA := len(chunksA)
	nChunksB := len(chunksB)

	for i := range chunksA {
		if i >= nChunksB {
			return false
		}

		aInt, aErr := strconv.Atoi(chunksA[i])
		bInt, bErr := strconv.Atoi(chunksB[i])

		// If both chunks are numeric, compare them as integers
		if aErr == nil && bErr == nil {
			if aInt == bInt {
				if i == nChunksA-1 {
					// We reached the last chunk of A, thus B is greater than A
					return true
				} else if i == nChunksB-1 {
					// We reached the last chunk of B, thus A is greater than B
					return false
				}
				continue
			}
			return aInt < bInt
		}
		// So far both strings are equal, continue to next chunk
		if chunksA[i] == chunksB[i] {
			if i == nChunksA-1 {
				// We reached the last chunk of A, thus B is greater than A
				return true
			} else if i == nChunksB-1 {
				// We reached the last chunk of B, thus A is greater than B
				return false
			}
			continue
		}
		return chunksA[i] < chunksB[i]
	}
	return true
}
