package gst

/*
#include <stdlib.h>
#include <gst/gst.h>
*/
import "C"
import "unsafe"

type Structure C.GstStructure

func (s *Structure) GetName() string {
	str := (*C.char)(C.gst_structure_get_name(s.g()))
	defer C.free(unsafe.Pointer(str))
	return C.GoString(str)
}
