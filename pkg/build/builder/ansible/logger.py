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


lib.LogLn.argtypes = [c_char_p]
lib.LogLn.restype = None

def LogLn(data):
    lib.LogLn(data)


lib.LogHighlightLn.argtypes = [c_char_p]
lib.LogHighlightLn.restype = None

def LogHighlightLn(data):
    lib.LogHighlightLn(data)


lib.LogServiceLn.argtypes = [c_char_p]
lib.LogServiceLn.restype = None

def LogServiceLn(data):
    lib.LogServiceLn(data)


lib.LogInfoLn.argtypes = [c_char_p]
lib.LogInfoLn.restype = None

def LogInfoLn(data):
    lib.LogInfoLn(data)


lib.LogErrorLn.argtypes = [c_char_p]
lib.LogErrorLn.restype = None

def LogErrorLn(data):
    lib.LogErrorLn(data)


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


def Log(msg):
    print msg

def LogHighlight(msg):
    print msg

def LogService(msg):
    print msg

def LogInfo(msg):
    print msg

def LogError(msg):
    print msg

def LogProcessStart(msg):
    print "LogProcessStart"
    print msg

def LogProcessEnd():
    print "LogProcessEnd"

def LogProcessFail():
    print "LogProcessFail"

def LogProcessEndStep(msg):
    print "LogProcessEndStep"
    print msg
