package gst

/*
#include <stdlib.h>
#include <gst/gst.h>
*/
import "C"
import "unsafe"

type CapsFeatures C.GstCapsFeatures

func (cf *CapsFeatures) g() *C.GstCapsFeatures {
	return (*C.GstCapsFeatures)(cf)
}

func (cf *CapsFeatures) Contains(feature string) bool {
	fs := (*C.gchar)(C.CString(feature))
	defer C.free(unsafe.Pointer(fs))
	return C.gst_caps_features_contains(cf.g(), fs) != 0
}
