from ctypes import *

lib = cdll.LoadLibrary("./pylogger.so")

class Error(Exception): pass


lib.Init.argtypes = []
lib.Init.restype = c_char_p

def Init():
    res = lib.Init()
    if res != None:
        raise Error(res)


lib.DisablePrettyLog.argtypes = []
lib.DisablePrettyLog.restype = None

def DisablePrettyLog():
    lib.DisablePrettyLog()


lib.EnableLogColor.argtypes = []
lib.EnableLogColor.restype = None

def EnableLogColor():
    lib.EnableLogColor()


lib.DisableLogColor.argtypes = []
lib.DisableLogColor.restype = None

def DisableLogColor():
    lib.DisableLogColor()


lib.IndentUp.argtypes = []
lib.IndentUp.restype = None

def IndentUp():
    lib.IndentUp()


lib.IndentDown.argtypes = []
lib.IndentDown.restype = None

def IndentDown():
    lib.IndentDown()


lib.OptionalLnModeOn.argtypes = []
lib.OptionalLnModeOn.restype = None

def OptionalLnModeOn():
    lib.OptionalLnModeOn()


lib.Log.argtypes = [c_char_p]
lib.Log.restype = None

def Log(data):
    lib.Log(data)


lib.LogHighlight.argtypes = [c_char_p]
lib.LogHighlight.restype = None

def LogHighlight(data):
    lib.LogHighlight(data)


lib.LogService.argtypes = [c_char_p]
lib.LogService.restype = None

def LogService(data):
    lib.LogService(data)


lib.LogInfo.argtypes = [c_char_p]
lib.LogInfo.restype = None

def LogInfo(data):
    lib.LogInfo(data)


lib.LogError.argtypes = [c_char_p]
lib.LogError.restype = None

def LogError(data):
    lib.LogError(data)


lib.LogProcessStart.argtypes = [c_char_p]
lib.LogProcessStart.restype = None

def LogProcessStart(msg):
    lib.LogProcessStart(msg)


lib.LogProcessEnd.argtypes = [c_bool]
lib.LogProcessEnd.restype = None

def LogProcessEnd(**kwargs):
    lib.LogProcessEnd(kwargs.get("without_log_optional_ln", False))


lib.LogProcessFail.argtypes = [c_bool]
lib.LogProcessFail.restype = None

def LogProcessFail(**kwargs):
    lib.LogProcessFail(kwargs.get("without_log_optional_ln", False))


lib.LogProcessStepEnd.argtypes = [c_char_p]
lib.LogProcessStepEnd.restype = None

def LogProcessStepEnd(msg):
    lib.LogProcessFail(msg)


lib.FitText.argtypes = [c_char_p, c_int, c_int, c_bool]
lib.FitText.restype = c_char_p

def FitText(text, **kwargs):
    return lib.FitText(text, kwargs.get("extra_indent_width", 0), kwargs.get("max_width", 0),kwargs.get("mark_wrapped_file", False))


lib.GetRawStreamsOutputMode.argtypes = []
lib.GetRawStreamsOutputMode.restype = c_bool

def GetRawStreamsOutputMode():
    return lib.GetRawStreamsOutputMode()


lib.RawStreamsOutputModeOn.argtypes = []
lib.RawStreamsOutputModeOn.restype = None

def RawStreamsOutputModeOn():
    lib.RawStreamsOutputModeOn()


lib.RawStreamsOutputModeOff.argtypes = []
lib.RawStreamsOutputModeOff.restype = None

def RawStreamsOutputModeOff():
    lib.RawStreamsOutputModeOff()


lib.MuteOut.argtypes = []
lib.MuteOut.restype = None

def MuteOut():
    lib.MuteOut()


lib.UnmuteOut.argtypes = []
lib.UnmuteOut.restype = None

def UnmuteOut():
    lib.UnmuteOut()


lib.MuteErr.argtypes = []
lib.MuteErr.restype = None

def MuteErr():
    lib.MuteErr()


lib.UnmuteErr.argtypes = []
lib.UnmuteErr.restype = None

def UnmuteErr():
    lib.UnmuteErr()
