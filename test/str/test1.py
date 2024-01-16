
def get_val():
    return 'xxx'


if __name__ == '__main__':
    decoyStr = 'blah'
    decoyBool = True
    decoyInt = 2
    decoyFloat = 2.2
    decoyStr2: str = 'blah'
    decoyBool2: bool = True
    decoyInt2: int = 2
    decoyFloat2: float = 2.2

    _s_: str = 'Hello'
    _s_ += ' World!!!'
    _s_ -= '!!'
    x = _s_
    _s_ = get_val()

    print(x)
