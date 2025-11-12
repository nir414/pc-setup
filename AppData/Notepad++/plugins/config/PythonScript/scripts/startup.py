def on_buffer_activated(args):
    editor.setTechnology(0)

notepad.callback(on_buffer_activated, [NOTIFICATION.BUFFERACTIVATED])
