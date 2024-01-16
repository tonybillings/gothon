package code

/*******************************************************************************
 bool
*******************************************************************************/

const boolSetFuncTemplate = `
def gothon_{{var_id}}_set(val: bool) -> bool:
    if val:
        _sock_{{var_id}}_set_in.send((1).to_bytes(1, 'big'))
    else:
        _sock_{{var_id}}_set_in.send((0).to_bytes(1, 'big'))
    _sock_{{var_id}}_set_out.recvfrom(1)
    return val`

const boolGetFuncTemplate = `
def gothon_{{var_id}}_get() -> bool:
    _sock_{{var_id}}_get_in.send((22).to_bytes(1, 'big'))
    val_bytes, _ = _sock_{{var_id}}_get_out.recvfrom(1)
    return val_bytes[0] != 0`

/*******************************************************************************
 int
*******************************************************************************/

const intSetFuncTemplate = `
def gothon_{{var_id}}_set(val: int) -> int:
    _sock_{{var_id}}_set_in.send(val.to_bytes(8, 'big', signed=True))
    _sock_{{var_id}}_set_out.recvfrom(1)
    return val`

const intGetFuncTemplate = `
def gothon_{{var_id}}_get() -> int:
    _sock_{{var_id}}_get_in.send((22).to_bytes(1, 'big'))
    val_bytes, _ = _sock_{{var_id}}_get_out.recvfrom(8)
    return int.from_bytes(val_bytes, 'big', signed=True)`

const intAddFuncTemplate = `
def gothon_{{var_id}}_add(delta: int):
    _sock_{{var_id}}_add_in.send(delta.to_bytes(8, 'big', signed=True))
    _sock_{{var_id}}_add_out.recvfrom(1)
    return delta`

const intSubFuncTemplate = `
def gothon_{{var_id}}_sub(delta: int):
    _sock_{{var_id}}_sub_in.send(delta.to_bytes(8, 'big', signed=True))
    _sock_{{var_id}}_sub_out.recvfrom(1)
    return delta`

const intMulFuncTemplate = `
def gothon_{{var_id}}_mul(multiplier: int):
    _sock_{{var_id}}_mul_in.send(multiplier.to_bytes(8, 'big', signed=True))
    _sock_{{var_id}}_mul_out.recvfrom(1)
    return multiplier`

const intDivFuncTemplate = `
def gothon_{{var_id}}_div(divisor: int):
    _sock_{{var_id}}_div_in.send(divisor.to_bytes(8, 'big', signed=True))
    _sock_{{var_id}}_div_out.recvfrom(1)
    return divisor`

/*******************************************************************************
 float
*******************************************************************************/

const floatSetFuncTemplate = `
def gothon_{{var_id}}_set(val: float) -> float:
    _sock_{{var_id}}_set_in.send(struct.pack('<d', val))
    _sock_{{var_id}}_set_out.recvfrom(1)
    return val`

const floatGetFuncTemplate = `
def gothon_{{var_id}}_get() -> float:
    _sock_{{var_id}}_get_in.send((22).to_bytes(1, 'little'))
    val_bytes, _ = _sock_{{var_id}}_get_out.recvfrom(8)
    return struct.unpack_from('<d', val_bytes, 0)[0]`

const floatAddFuncTemplate = `
def gothon_{{var_id}}_add(delta: float):
    _sock_{{var_id}}_add_in.send(bytearray(struct.pack('<d', delta)))
    _sock_{{var_id}}_add_out.recvfrom(1)
    return delta`

const floatSubFuncTemplate = `
def gothon_{{var_id}}_sub(delta: float):
    _sock_{{var_id}}_sub_in.send(bytearray(struct.pack('<d', delta)))
    _sock_{{var_id}}_sub_out.recvfrom(1)
    return delta`

const floatMulFuncTemplate = `
def gothon_{{var_id}}_mul(multiplier: float):
    _sock_{{var_id}}_mul_in.send(bytearray(struct.pack('<d', multiplier)))
    _sock_{{var_id}}_mul_out.recvfrom(1)
    return multiplier`

const floatDivFuncTemplate = `
def gothon_{{var_id}}_div(divisor: float):
    _sock_{{var_id}}_div_in.send(bytearray(struct.pack('<d', divisor)))
    _sock_{{var_id}}_div_out.recvfrom(1)
    return divisor`

/*******************************************************************************
 str
*******************************************************************************/

const stringSetFuncTemplate = `
def gothon_{{var_id}}_set(val: str) -> str:
    _sock_{{var_id}}_set_in.send(bytes(val, 'utf-8'))
    _sock_{{var_id}}_set_out.recvfrom(1)
    return val`

const stringGetFuncTemplate = `
def gothon_{{var_id}}_get() -> str:
    _sock_{{var_id}}_get_in.send((22).to_bytes(1, 'big'))
    val_bytes, _ = _sock_{{var_id}}_get_out.recvfrom({{str_max_size}})
    return str(val_bytes, 'utf-8')`

const stringAddFuncTemplate = `
def gothon_{{var_id}}_add(suffix: str):
    _sock_{{var_id}}_add_in.send(bytes(suffix, 'utf-8'))
    _sock_{{var_id}}_add_out.recvfrom(1)
    return suffix`

const stringSubFuncTemplate = `
def gothon_{{var_id}}_sub(suffix: str):
    _sock_{{var_id}}_sub_in.send(bytes(suffix, 'utf-8'))
    _sock_{{var_id}}_sub_out.recvfrom(1)
    return suffix`

/*******************************************************************************
 mutex
*******************************************************************************/

const mutexFuncTemplate = `
def gothon_{{var_id}}():
    _sock_{{var_id}}_in.send((22).to_bytes(1, 'big'))
    _sock_{{var_id}}_out.recvfrom(1)`

/*******************************************************************************
 sync
*******************************************************************************/

const syncFuncTemplate = `
def gothon_{{var_id}}(n: int = 0):
    _sock_{{var_id}}_in.send(n.to_bytes(4, 'big'))
    _sock_{{var_id}}_out.recvfrom(1)`

/*******************************************************************************
 queue
*******************************************************************************/

const queueSizeFuncTemplate = `
def gothon_{{var_id}}_size() -> int:
    _sock_{{var_id}}_size_in.send((22).to_bytes(1, 'big'))
    val_bytes, _ = _sock_{{var_id}}_size_out.recvfrom(8)
    return int.from_bytes(val_bytes, 'big', signed=False)`

const queueEmptyFuncTemplate = `
def gothon_{{var_id}}_empty() -> bool:
    _sock_{{var_id}}_empty_in.send((22).to_bytes(1, 'big'))
    val_bytes, _ = _sock_{{var_id}}_empty_out.recvfrom(1)
    return val_bytes[0] != 0`

const queueFullFuncTemplate = `
def gothon_{{var_id}}_full() -> bool:
    _sock_{{var_id}}_full_in.send((22).to_bytes(1, 'big'))
    val_bytes, _ = _sock_{{var_id}}_full_out.recvfrom(1)
    return val_bytes[0] != 0`

/*******************************************************************************
 bool queue
*******************************************************************************/

const boolQueueSetFuncTemplate = `
def gothon_{{var_id}}_set(val: bool) -> (bool, bool):
    if val:
        _sock_{{var_id}}_set_in.send((1).to_bytes(1, 'big'))
    else:
        _sock_{{var_id}}_set_in.send((0).to_bytes(1, 'big'))
    ok, _ = _sock_{{var_id}}_set_out.recvfrom(1)
    return val, ok[0] == 22`

const boolQueueGetFuncTemplate = `
def gothon_{{var_id}}_get() -> (bool, bool):
    _sock_{{var_id}}_get_in.send((22).to_bytes(1, 'big'))
    ok, _ = _sock_{{var_id}}_get_ok.recvfrom(1)
    if ok[0] == 22:
        val_bytes, _ = _sock_{{var_id}}_get_out.recvfrom(1)
        return val_bytes[0] != 0, True
    else:
        return False, False`

/*******************************************************************************
 int queue
*******************************************************************************/

const intQueueSetFuncTemplate = `
def gothon_{{var_id}}_set(val: int) -> (int, bool):
    _sock_{{var_id}}_set_in.send(val.to_bytes(8, 'big', signed=True))
    ok, _ = _sock_{{var_id}}_set_out.recvfrom(1)
    return val, ok[0] == 22`

const intQueueGetFuncTemplate = `
def gothon_{{var_id}}_get() -> (int, bool):
    _sock_{{var_id}}_get_in.send((22).to_bytes(1, 'big'))
    ok, _ = _sock_{{var_id}}_get_ok.recvfrom(1)
    if ok[0] == 22:
        val_bytes, _ = _sock_{{var_id}}_get_out.recvfrom(8)
        return int.from_bytes(val_bytes, 'big', signed=True), True
    else:
        return 0, False`

/*******************************************************************************
 float queue
*******************************************************************************/

const floatQueueSetFuncTemplate = `
def gothon_{{var_id}}_set(val: float) -> (float, bool):
    _sock_{{var_id}}_set_in.send(struct.pack('<d', val))
    ok, _ = _sock_{{var_id}}_set_out.recvfrom(1)
    return val, ok[0] == 22`

const floatQueueGetFuncTemplate = `
def gothon_{{var_id}}_get() -> (float, bool):
    _sock_{{var_id}}_get_in.send((22).to_bytes(1, 'little'))
    ok, _ = _sock_{{var_id}}_get_ok.recvfrom(1)
    if ok[0] == 22:
        val_bytes, _ = _sock_{{var_id}}_get_out.recvfrom(8)
        return struct.unpack_from('<d', val_bytes, 0)[0], True
    else:
        return 0, False`

/*******************************************************************************
 string queue
*******************************************************************************/

const stringQueueSetFuncTemplate = `
def gothon_{{var_id}}_set(val: str) -> (str, bool):
    _sock_{{var_id}}_set_in.send(bytes(val, 'utf-8'))
    ok, _ = _sock_{{var_id}}_set_out.recvfrom(1)
    return val, ok[0] == 22`

const stringQueueGetFuncTemplate = `
def gothon_{{var_id}}_get() -> (str, bool):
    _sock_{{var_id}}_get_in.send((22).to_bytes(1, 'big'))
    ok, _ = _sock_{{var_id}}_get_ok.recvfrom(1)
    if ok[0] == 22:
        val_bytes, _ = _sock_{{var_id}}_get_out.recvfrom({{str_max_size}})
        return str(val_bytes, 'utf-8'), True
    else:
        "", False`

/*******************************************************************************
 socket
*******************************************************************************/

const socketInitTemplate = `
    _sock_{{var_id}}_{{action}}_in.connect(_addr_{{var_id}}_{{action}}_in)
    _sock_{{var_id}}_{{action}}_out.bind(_addr_{{var_id}}_{{action}}_out)`

const socketInitTemplateForMutex = `
    _sock_{{var_id}}_in.connect(_addr_{{var_id}}_in)
    _sock_{{var_id}}_out.bind(_addr_{{var_id}}_out)`

const socketInitTemplateForSync = `
    _sock_{{var_id}}_in.connect(_addr_{{var_id}}_in)
    _sock_{{var_id}}_out.bind(_addr_{{var_id}}_out)`

const socketInitTemplateForQueueGet = `
    _sock_{{var_id}}_get_ok.bind(_addr_{{var_id}}_get_ok)`

/*******************************************************************************
 template map
*******************************************************************************/

var templates = map[string]string{
	"bool_set":        boolSetFuncTemplate,
	"bool_get":        boolGetFuncTemplate,
	"int_set":         intSetFuncTemplate,
	"int_get":         intGetFuncTemplate,
	"int_add":         intAddFuncTemplate,
	"int_sub":         intSubFuncTemplate,
	"int_mul":         intMulFuncTemplate,
	"int_div":         intDivFuncTemplate,
	"float_set":       floatSetFuncTemplate,
	"float_get":       floatGetFuncTemplate,
	"float_add":       floatAddFuncTemplate,
	"float_sub":       floatSubFuncTemplate,
	"float_mul":       floatMulFuncTemplate,
	"float_div":       floatDivFuncTemplate,
	"str_set":         stringSetFuncTemplate,
	"str_get":         stringGetFuncTemplate,
	"str_add":         stringAddFuncTemplate,
	"str_sub":         stringSubFuncTemplate,
	"mutex":           mutexFuncTemplate,
	"sync":            syncFuncTemplate,
	"queue_size":      queueSizeFuncTemplate,
	"queue_empty":     queueEmptyFuncTemplate,
	"queue_full":      queueFullFuncTemplate,
	"bool_queue_set":  boolQueueSetFuncTemplate,
	"bool_queue_get":  boolQueueGetFuncTemplate,
	"int_queue_set":   intQueueSetFuncTemplate,
	"int_queue_get":   intQueueGetFuncTemplate,
	"float_queue_set": floatQueueSetFuncTemplate,
	"float_queue_get": floatQueueGetFuncTemplate,
	"str_queue_set":   stringQueueSetFuncTemplate,
	"str_queue_get":   stringQueueGetFuncTemplate,
}
