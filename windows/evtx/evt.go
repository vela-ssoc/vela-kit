package evtx

import (
	"fmt"
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

/*Functionality related to events and listening to the vela-event log*/

// Get a handle to a render context which will render properties from the System element.
//
//	Wraps EvtCreateRenderContext() with Flags = EvtRenderContextSystem. The resulting
//	handle must be closed with CloseEventHandle.
func GetSystemRenderContext() (SysRenderContext, error) {
	context, err := EvtCreateRenderContext(0, 0, EvtRenderContextSystem)
	if err != nil {
		return 0, err
	}
	return SysRenderContext(context), nil
}

/*
Get a handle for a vela-event log subscription on the given channel.

	`query` is an XPath expression to filter the events on the channel - "*" allows all events.
	The resulting handle must be closed with CloseEventHandle.
*/
func CreateListener(channel, query string, startpos EVT_SUBSCRIBE_FLAGS, watcher *LogEventCallbackWrapper) (ListenerHandle, error) {
	wideChan, err := syscall.UTF16PtrFromString(channel)
	if err != nil {
		return 0, err
	}
	wideQuery, err := syscall.UTF16PtrFromString(query)
	if err != nil {
		return 0, err
	}
	listenerHandle, err := EvtSubscribe(0, 0, wideChan, wideQuery, 0, uintptr(unsafe.Pointer(watcher)), uintptr(syscall.NewCallback(eventCallback)), uint32(startpos))
	if err != nil {
		return 0, err
	}
	return ListenerHandle(listenerHandle), nil
}

/*
Get a handle for an vela-event log subscription on the given channel. Will begin at the

	bookmarked vela-event, or the closest possible vela-event if the log has been truncated.
	`query` is an XPath expression to filter the events on the channel - "*" allows all events.
	The resulting handle must be closed with CloseEventHandle.
*/
func CreateListenerFromBookmark(channel, query string, watcher *LogEventCallbackWrapper, bookmarkHandle BookmarkHandle) (ListenerHandle, error) {
	wideChan, err := syscall.UTF16PtrFromString(channel)
	if err != nil {
		return 0, err
	}
	wideQuery, err := syscall.UTF16PtrFromString(query)
	if err != nil {
		return 0, err
	}
	listenerHandle, err := EvtSubscribe(0, 0, wideChan, wideQuery, syscall.Handle(bookmarkHandle), uintptr(unsafe.Pointer(watcher)), syscall.NewCallback(eventCallback), uint32(EvtSubscribeStartAfterBookmark))
	if err != nil {
		return 0, err
	}
	return ListenerHandle(listenerHandle), nil
}

/* Get the formatted string that represents this message. This method wraps EvtFormatMessage. */
func FormatMessage(eventPublisherHandle PublisherHandle, eventHandle EventHandle, format EVT_FORMAT_MESSAGE_FLAGS) (string, error) {
	var size uint32 = 0
	err := EvtFormatMessage(syscall.Handle(eventPublisherHandle), syscall.Handle(eventHandle), 0, 0, nil, uint32(format), 0, nil, &size)
	if err != nil {
		if errno, ok := err.(syscall.Errno); !ok || errno != 122 {
			// Check if the error is ERR_INSUFICIENT_BUFFER
			return "", err
		}
	}
	buf := make([]uint16, size)
	err = EvtFormatMessage(syscall.Handle(eventPublisherHandle), syscall.Handle(eventHandle), 0, 0, nil, uint32(format), uint32(len(buf)), &buf[0], &size)
	if err != nil {
		return "", err
	}
	return syscall.UTF16ToString(buf), nil
}

/* Get the formatted string for the last error which occurred. Wraps GetLastError and FormatMessage. */
func GetLastError() error {
	return syscall.GetLastError()
}

/*
Render the system properties from the vela-event and returns an array of properties.

	Properties can be accessed using RenderStringField, RenderIntField, RenderFileTimeField,
	or RenderUIntField depending on type. This buffer must be freed after use.
*/
func RenderEventValues(renderContext SysRenderContext, eventHandle EventHandle) (EvtVariant, error) {
	var bufferUsed uint32 = 0
	var propertyCount uint32 = 0
	err := EvtRender(syscall.Handle(renderContext), syscall.Handle(eventHandle), EvtRenderEventValues, 0, nil, &bufferUsed, &propertyCount)
	if bufferUsed == 0 {
		return nil, err
	}
	buffer := make([]byte, bufferUsed)
	bufSize := bufferUsed
	err = EvtRender(syscall.Handle(renderContext), syscall.Handle(eventHandle), EvtRenderEventValues, bufSize, (*uint16)(unsafe.Pointer(&buffer[0])), &bufferUsed, &propertyCount)
	if err != nil {
		return nil, err
	}
	return NewEvtVariant(buffer), nil
}

// Render the vela-event as XML.
func RenderEventXML(eventHandle EventHandle) (string, error) {
	var bufferUsed, propertyCount uint32

	err := EvtRender(0, syscall.Handle(eventHandle), EvtRenderEventXml, 0, nil, &bufferUsed, &propertyCount)

	if bufferUsed == 0 {
		return "", err
	}

	buffer := make([]uint16, bufferUsed)
	bufSize := bufferUsed

	err = EvtRender(0, syscall.Handle(eventHandle), EvtRenderEventXml, bufSize, &buffer[0], &bufferUsed, &propertyCount)
	if err != nil {
		return err.Error(), err
	}

	// Remove null bytes
	//xml := bytes.Replace(buffer, []byte("\x00"), []byte{}, -1)

	//xml16 := (*uint16)(unsafe.Pointer(&buffer[0]))
	xml := windows.UTF16ToString(buffer)

	return xml, nil
}

/* Get a handle that represents the publisher of the vela-event, given the rendered vela-event values. */
func GetEventPublisherHandle(renderedFields EvtVariant) (PublisherHandle, error) {
	publisher, err := renderedFields.String(EvtSystemProviderName)
	if err != nil {
		return 0, err
	}
	widePublisher, err := syscall.UTF16PtrFromString(publisher)
	if err != nil {
		return 0, err
	}
	handle, err := EvtOpenPublisherMetadata(0, widePublisher, nil, 0, 0)
	if err != nil {
		return 0, err
	}
	return PublisherHandle(handle), nil
}

/* Close an vela-event handle. */
func CloseEventHandle(handle uint64) error {
	return EvtClose(syscall.Handle(handle))
}

/* Cancel pending actions on the vela-event handle. */
func CancelEventHandle(handle uint64) error {
	err := EvtCancel(syscall.Handle(handle))
	if err != nil {
		return err
	}
	return nil
}

/* Get the first vela-event in the log, for testing */
func getTestEventHandle() (EventHandle, error) {
	wideQuery, _ := syscall.UTF16PtrFromString("*")
	wideChannel, _ := syscall.UTF16PtrFromString("Application")
	handle, err := EvtQuery(0, wideChannel, wideQuery, EvtQueryChannelPath)
	if err != nil {
		return 0, err
	}

	var record syscall.Handle
	var recordsReturned uint32
	err = EvtNext(handle, 1, &record, 500, 0, &recordsReturned)
	if err != nil {
		EvtClose(handle)
		return 0, nil
	}
	EvtClose(handle)
	return EventHandle(record), nil
}

func eventCallback(Action uint32, Context unsafe.Pointer, handle syscall.Handle) uintptr {
	cbWrap := (*LogEventCallbackWrapper)(Context)
	if Action == 0 {
		cbWrap.callback.PublishError(fmt.Errorf("Event log callback got error: %v", GetLastError()))
	} else {
		cbWrap.callback.PublishEvent(EventHandle(handle), cbWrap.subscribedChannel)
	}
	return 0
}

// CreateMap converts the WinLogEvent to a map[string]interface{}
func (ev *WinLogEvent) CreateMap() map[string]interface{} {
	ret := make(map[string]interface{})
	ret["XmlText"] = ev.XmlText
	ret["ProviderName"] = ev.ProviderName
	ret["EventId"] = ev.EventId
	ret["Qualifiers"] = ev.Qualifiers
	ret["Level"] = ev.Level
	ret["Task"] = ev.Task
	ret["Opcode"] = ev.Opcode
	ret["Created"] = ev.Created
	ret["RecordId"] = ev.RecordId
	ret["ProcessId"] = ev.ProcessId
	ret["ThreadId"] = ev.ThreadId
	ret["Channel"] = ev.Channel
	ret["ComputerName"] = ev.ComputerName
	ret["Version"] = ev.Version
	ret["Msg"] = ev.Msg
	ret["LevelText"] = ev.LevelText
	ret["TaskText"] = ev.TaskText
	ret["OpcodeText"] = ev.OpcodeText
	ret["Keywords"] = ev.Keywords
	ret["ChannelText"] = ev.ChannelText
	ret["ProviderText"] = ev.ProviderText
	ret["IdText"] = ev.IdText
	ret["Bookmark"] = ev.Bookmark
	ret["SubscribedChannel"] = ev.SubscribedChannel
	ret["Bookmark"] = ev.Bookmark
	return ret
}
