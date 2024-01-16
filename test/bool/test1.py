
def get_val():
    return False


if __name__ == '__main__':
    decoyStr = 'blah'
    decoyBool = True
    decoyInt = 2
    decoyFloat = 2.2
    decoyStr2: str = 'blah'
    decoyBool2: bool = True
    decoyInt2: int = 2
    decoyFloat2: float = 2.2

    _b_: bool = False
    _b_ = True
    x = False
    if _b_:
        x = _b_
        _b_ = get_val()

    print(x)
