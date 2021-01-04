package gst

/*
#include <stdlib.h>
#include <gst/gst.h>

typedef struct PadAddedData {
  unsigned int signalElementPadAddedId;
  gpointer data;
} t_PadAddedData;

static t_PadAddedData * makePadAddedData(unsigned int id, gpointer data) {
	t_PadAddedData * ret = malloc(sizeof(t_PadAddedData));
	ret->signalElementPadAddedId = id;
	ret->data = data;
	return ret;
}

//void deletePadAddedData(t_PadAddedData * data) {
//	free(data->data);
//	free(data);
//}

void element_padAddedHandler(GstElement * element, GstPad * pad, gpointer data);

static gulong Element_signal_connect_pad_added(gpointer instance, gpointer data) {
	return g_signal_connect(instance, "pad-added", G_CALLBACK(element_padAddedHandler), data);
}

*/
import "C"

import (
	"github.com/ziutek/glib"
	"sync"
	"unsafe"
)

type State C.GstState

const (
	STATE_VOID_PENDING = State(C.GST_STATE_VOID_PENDING)
	STATE_NULL         = State(C.GST_STATE_NULL)
	STATE_READY        = State(C.GST_STATE_READY)
	STATE_PAUSED       = State(C.GST_STATE_PAUSED)
	STATE_PLAYING      = State(C.GST_STATE_PLAYING)
)

func (s *State) g() *C.GstState {
	return (*C.GstState)(s)
}

func (s State) String() string {
	switch s {
	case STATE_VOID_PENDING:
		return "STATE_VOID_PENDING"
	case STATE_NULL:
		return "STATE_NULL"
	case STATE_READY:
		return "STATE_READY"
	case STATE_PAUSED:
		return "STATE_PAUSED"
	case STATE_PLAYING:
		return "STATE_PLAYING"
	}
	panic("Unknown state")
}

type StateChangeReturn C.GstStateChangeReturn

const (
	STATE_CHANGE_FAILURE    = StateChangeReturn(C.GST_STATE_CHANGE_FAILURE)
	STATE_CHANGE_SUCCESS    = StateChangeReturn(C.GST_STATE_CHANGE_SUCCESS)
	STATE_CHANGE_ASYNC      = StateChangeReturn(C.GST_STATE_CHANGE_ASYNC)
	STATE_CHANGE_NO_PREROLL = StateChangeReturn(C.GST_STATE_CHANGE_NO_PREROLL)
)


type Element struct {
	GstObj
}

func (e *Element) g() *C.GstElement {
	return (*C.GstElement)(e.GetPtr())
}

func (e *Element) AsElement() *Element {
	return e
}

func (e *Element) Link(next ...*Element) bool {
	for _, dst := range next {
		if C.gst_element_link(e.g(), dst.g()) == 0 {
			return false
		}
		e = dst
	}
	return true
}

func (e *Element) Unlink(next ...*Element) {
	for _, dst := range next {
		C.gst_element_unlink(e.g(), dst.g())
		e = dst
	}
}

func (e *Element) LinkFiltered(dst *Element, filter *Caps) bool {
	return C.gst_element_link_filtered(e.g(), dst.g(), filter.g()) != 0
}

func (e *Element) LinkPads(pad_name string, dst *Element, dst_pad_name string) bool {
	src_pname := (*C.gchar)(C.CString(pad_name))
	defer C.free(unsafe.Pointer(src_pname))
	dst_pname := (*C.gchar)(C.CString(dst_pad_name))
	defer C.free(unsafe.Pointer(dst_pname))
	return C.gst_element_link_pads(e.g(), src_pname, dst.g(), dst_pname) != 0
}

func (e *Element) UnlinkPads(pad_name string, dst *Element, dst_pad_name string) {
	src_pname := (*C.gchar)(C.CString(pad_name))
	defer C.free(unsafe.Pointer(src_pname))
	dst_pname := (*C.gchar)(C.CString(dst_pad_name))
	defer C.free(unsafe.Pointer(dst_pname))
	C.gst_element_unlink_pads(e.g(), src_pname, dst.g(), dst_pname)
}

func (e *Element) SetState(state State) StateChangeReturn {
	return StateChangeReturn(C.gst_element_set_state(e.g(), C.GstState(state)))
}

func (e *Element) GetState(timeout_ns int64) (state, pending State,
	ret StateChangeReturn) {
	ret = StateChangeReturn(C.gst_element_get_state(
		e.g(), state.g(), pending.g(), C.GstClockTime(timeout_ns),
	))
	return
}

func (e *Element) AddPad(p *Pad) bool {
	return C.gst_element_add_pad(e.g(), p.g()) != 0
}

func (e *Element) GetRequestPad(name string) *Pad {
	s := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(s))
	cp := C.gst_element_get_request_pad(e.g(), s)
	if cp == nil {
		return nil
	}
	p := new(Pad)
	p.SetPtr(glib.Pointer(cp))
	return p
}

func (e *Element) GetStaticPad(name string) *Pad {
	s := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(s))
	cp := C.gst_element_get_static_pad(e.g(), s)
	if cp == nil {
		return nil
	}
	p := new(Pad)
	p.SetPtr(glib.Pointer(cp))
	return p
}

func (e *Element) GetBus() *Bus {
	bus := C.gst_element_get_bus(e.g())
	if bus == nil {
		return nil
	}
	b := new(Bus)
	b.SetPtr(glib.Pointer(bus))
	return b
}


type signalElementPadAddedDetail struct {
	callback  ElementPadAddedCallback
	handlerID C.gulong
}

var signalElementPadAddedId int
var signalElementPadAddedMap = make(map[int]signalElementPadAddedDetail)
var signalElementPadAddedLock sync.RWMutex

// ElementPadAddedCallback is a callback function for a 'pad-added' signal emitted from a Selection.
type ElementPadAddedCallback func(element *Element, srcPad *Pad, data glib.Pointer)

/*
ConnectPadAdded connects the callback to the 'pad-added' signal for the Selection.
The returned value represents the connection, and may be passed to DisconnectPadAdded to remove it.
*/
func (e *Element) ConnectPadAdded(callback ElementPadAddedCallback, data glib.Pointer) int {
	signalElementPadAddedLock.Lock()
	defer signalElementPadAddedLock.Unlock()

	signalElementPadAddedId++

	d := C.makePadAddedData(C.uint(signalElementPadAddedId), C.gpointer(data))

	instance := e.g()
	handlerID := C.Element_signal_connect_pad_added(C.gpointer(instance), C.gpointer(d))

	detail := signalElementPadAddedDetail{callback, handlerID}
	signalElementPadAddedMap[signalElementPadAddedId] = detail

	return signalElementPadAddedId
}

/*
DisconnectPadAdded disconnects a callback from the 'pad-added' signal for the Selection.
The connectionID should be a value returned from a call to ConnectPadAdded.
*/
func (e *Element) DisconnectPadAdded(connectionID int) {
	signalElementPadAddedLock.Lock()
	defer signalElementPadAddedLock.Unlock()

	detail, exists := signalElementPadAddedMap[connectionID]
	if !exists {
		return
	}

	instance := e.g()
	C.g_signal_handler_disconnect(C.gpointer(instance), detail.handlerID)
	delete(signalElementPadAddedMap, connectionID)
}

//export element_padAddedHandler
func element_padAddedHandler(element *C.GstElement, pad *C.GstPad, data C.gpointer) {
	signalElementPadAddedLock.RLock()
	defer signalElementPadAddedLock.RUnlock()

	e := new(Element)
	e.SetPtr(glib.Pointer(element))

	p := new(Pad)
	p.SetPtr(glib.Pointer(pad))

	d := (*C.t_PadAddedData)(data)

	index := int(d.signalElementPadAddedId)
	callback := signalElementPadAddedMap[index].callback
	callback(e, p, glib.Pointer(d.data))
}

// TODO: Move ElementFactoryMake to element_factory.go
func ElementFactoryMake(factory_name, name string) *Element {
	fn := (*C.gchar)(C.CString(factory_name))
	defer C.free(unsafe.Pointer(fn))
	n := (*C.gchar)(C.CString(name))
	defer C.free(unsafe.Pointer(n))
	ge := C.gst_element_factory_make(fn, n)
	if ge == nil {
		return nil
	}
	e := new(Element)
	e.SetPtr(glib.Pointer(ge))
	return e
}
