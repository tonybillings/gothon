# Gothon

## Overview

Gothon enables ***Python*** developers the ability to easily parallelize their application, writing performant, concurrent code without having to switch interpreters* or import some module and learn another API!  Using only standard Python syntax, you can use Gothon to launch multiple instances of your script/application simultaneously, access shared memory resources concurrently, leverage mutexes and wait groups, and assign/distribute work to specific application instances (**nodes** in Gothon lexicon).

<sup>* Gothon is 1 part Python interpreter, 99 parts Python wrapper!  It invokes whatever interpreter is mapped to the alias `python`, but before doing so must translate operations relating to the variables it manages.</sup>

## The GIL Problem

Developers coming from other languages may be surprised when they first learn about the Global Interpreter Lock, or GIL, which essentially makes simultaneous execution of threads impossible.  If the threads are waiting for some external resource (is IO-bound), then it's not a problem but for CPU-bound tasks it very much is.  There are many articles on the Internet that go into this issue in greater detail, however an example will follow here as a means of introducing and contrasting Gothon.       

Consider this contrived example:
```python
import time

def do_work_1(n):
    for i in range(n):
        n * n / n % n + n - n * i

def do_work_2(n):
    for i in range(n):
        n * n / n % n + n - n * i

if __name__ == "__main__":
    start_time = time.time()
    do_work_1(10_000_000)
    do_work_2(10_000_000)
    end_time = time.time()
    run_time = end_time - start_time
    print(f'runtime: {run_time}')
```

As you can see, the two functions do the exact same work and the application is clearly CPU-bound.  The functions are also invoked sequentially and everything is running on the main thread.  

The time it takes for this script to run on a test system is, on average, **4.12 seconds**.

Since the two functions are not dependent on each other they can be executed in parallel.  Here's the same script but rewritten to use the `threading` module.
```python
import threading
import time

def do_work_1(n):
    for i in range(n):
        n * n / n % n + n - n * i

def do_work_2(n):
    for i in range(n):
        n * n / n % n + n - n * i

if __name__ == "__main__":
    start_time = time.time()
    t1 = threading.Thread(target=do_work_1, args=(10_000_000,))
    t2 = threading.Thread(target=do_work_2, args=(10_000_000,))
    t1.start()
    t2.start()
    t1.join()
    t2.join()
    end_time = time.time()
    run_time = end_time - start_time
    print(f'runtime: {run_time}')
```

Naive developers may look at this code and be certain that the two functions will be executed in parallel, thus reducing the runtime by some amount; in fact, it increased to **4.23 seconds**!  Since the work can't be executed in parallel, all it ended up doing is tasking the CPU with more work, as it now has to manage two new threads (and with only one of them actually ending up doing any work).

Since the GIL prevents multiple threads within a process from executing simultaneously, the only solution is to employ a multiprocessing paradigm and this is where Gothon comes into play...as well as the standard Python module `multiprocessing`.  Before we see how Gothon attacks the problem, let's see the previous example rewritten to be truly parallelized, using the `multiprocessing` module:
```python
from multiprocessing import Process
import time

def do_work_1(n):
    for i in range(n):
        n * n / n % n + n - n * i

def do_work_2(n):
    for i in range(n):
        n * n / n % n + n - n * i

if __name__ == "__main__":
    start_time = time.time()
    p1 = Process(target=do_work_1, args=([10_000_000]))
    p2 = Process(target=do_work_2, args=([10_000_000]))
    p1.start()
    p2.start()
    p1.join()
    p2.join()
    end_time = time.time()
    run_time = end_time - start_time
    print(f'runtime: {run_time}')
```

Pretty straight-forward, and as to be expected, the runtime was reduced by about half, to **2.11** seconds (average over 5 runs).

Now let's see it written in Gothon:
```python
import time

_node_: int = 0

def do_work_1(n):
    for i in range(n):
        n * n / n % n + n - n * i
    _sync_nodes_(1)

def do_work_2(n):
    for i in range(n):
        n * n / n % n + n - n * i
    _sync_nodes_(1)

if __name__ == "__main__":
    _sync_nodes_: callable = lambda n=2: ()

    start_time = time.time()

    if _node_ == 0:
        do_work_1(10_000_000)
    elif _node_ == 1:
        do_work_2(10_000_000)
        _sync_nodes_()
        end_time = time.time()
        run_time = end_time - start_time
        print(f'runtime: {run_time}')

```

As promised, there is no module to import so no *API* to learn.  Instead, Gothon relies on an "opinionated" approach, meaning it simply requires you to declare your Gothon-managed variables in a particular way...that you must learn, but that you can also customize!  :)   

In this example, there are two such variables, `_node_`, which is a special **system** variable, and `_sync_nodes_`, which is a regular **user** variable.  System and user variables will be explained later, but for now just know that running this script via `gothon 2 main` (assuming the file is named `main.py`) will result in a runtime of **2.05 seconds**!  The work was parallelized with virtually no overhead.

At this point, you may be thinking with such a small performance gain and equally simple syntax, why bother with Gothon?  That's a good question to ask and if your use case is as simple as this example then look no further than the `multiprocessing` module.  Things change dramatically, however, when you need to share data/state between the processes and when performance is paramount.    

Consider this example:
```python
import time

def increment_to_limit(n, limit):
    while n < limit:
        n += 1
    return n

if __name__ == "__main__":
    x = 0
    start_time = time.time()
    x = increment_to_limit(x, 1_000_000)
    end_time = time.time()
    run_time = end_time - start_time
    print(f'runtime: {run_time}')
```

The point is to increment as fast as it can until it reaches some threshold.  When single-threaded like this, it only takes **0.05 seconds** to complete, and it's easy to see why.  Parallelizing this wouldn't make any sense in the real world, but here it does so that we can focus on comparing performance with respect to concurrent variable access (and in other cases, execution synchronization).  Here's that script rewritten to use the `multiprocessing` module:
```python
from multiprocessing import Process, Manager
import time

def increment_to_limit(n, limit):
    while n.value < limit:
        n.value += 1

if __name__ == "__main__":
    x = Manager().Value('i', 0)
    start_time = time.time()
    p1 = Process(target=increment_to_limit, args=(x, 1_000_000))
    p2 = Process(target=increment_to_limit, args=(x, 1_000_000))
    p1.start()
    p2.start()
    p1.join()
    p2.join()
    end_time = time.time()
    run_time = end_time - start_time
    print(f'runtime: {run_time}')
```

...and here it is written in Gothon:

```python
import time

_node_: int = 0
_node_count_: int = 0

def increment_to_limit(_x_, limit):
    while _x_ < limit:
        _x_ += 1

if __name__ == "__main__":
    _x_: int = 0
    _sync_nodes_: callable = lambda n=_node_count_: ()

    start_time = time.time()

    increment_to_limit(_x_, 1_000_000)
    _sync_nodes_(1) # signal to other nodes this one is done working
    _sync_nodes_() # wait for other nodes to finish their work
    
    end_time = time.time()
    run_time = end_time - start_time
    print(f'runtime: {run_time}')
```

Whether you find one API/syntax more appealing versus another is of course a matter of preference, but in terms of performance, there is a clear winner:  

|             | 1 Proc/Node | 2 Procs/Nodes | 3 Procs/Nodes | 4 Procs/Nodes |
|-------------|------------:|--------------:|--------------:|--------------:|
| **CPython** |      107.80 |        130.11 |        183.48 |        229.22 |
| **Gothon**  |       21.72 |         12.86 |         11.54 |          9.96 |
<sub>Runtimes shown are in **seconds**.</sub>

The `multiprocessing` module has other features that may allow you to improve performance, but the complexity of going that route will quickly surpass that of using Gothon.  To be fair, this contrived example is not like many real-world scenarios, where the contention for shared memory is not so significant/frequent.

## How Gothon Works

Since multithreading in CPython is limited due to the GIL, the most common solution is to leverage another technique known as multiprocessing.  When your script/application is invoked via the `gothon` command instead of `python`, multiple instances will run simultaneously (in parallel), similar to having multiple VM's or containers within a cluster run the same application (horizontal scaling).  


Gothon is the backplane that manages these application instances-- or what it refers to as **nodes**.  The number of nodes that get instantiated is specified as the second argument to the `gothon` command.  For the so-called **user** and **system** variables that Gothon manages, they are shared/accessible by all nodes, each protected by their own mutex to ensure concurrency and thread-safety. 

Before your application can be executed, Gothon must first translate the code that changes/accesses the variables it manages into function calls that pass data to/from the backplane so that it can do that work on behalf of the node.  This inter-process communication (IPC) is done over Unix Domain Sockets (UDS) using datagrams.  Note that your original source code files never get modified, but instead a hidden folder named `.gothon` is created at the project root and the source is copied to that folder, one copy per node, with each node getting a customized version of a Gothon-created module that provides the UDS glue needed for transferring variable values as well as to communicate synchronization actions like waiting for a mutex unlock.


## Installation

If you already have the compiled binary for your architecture, then simply place it somewhere and ensure it's in your `PATH`.  

To compile/install from source, simply run the `install.sh` script located at this project's root directory.  The only prerequisite is that the Go compiler is installed (and in your PATH), and you can find instructions for that [here](https://go.dev/doc/install).

When running Gothon, it invokes whatever interpreter is mapped to the alias `python`, so be sure that is set correctly for all shells (specifically, non-login, non-interactive); same goes for the `go` command/alias!


## Command Usage

The syntax for the `gothon` command is:  `gothon NODE_COUNT MODULE_NAME [MODULE_ARG...]`

For example, if you have a single script to execute and its name is `main.py` and you want 8 instances of it to run simultaneously, then the command becomes:
```shell
gothon 8 main
```

This of course assumes you are running it from the same directory that contains `main.py`.

If your script accepts command-line arguments, they can simply follow the module name as they normally would.

Note that creating more nodes than you have processor cores (or hyper-threads / virtual cores) generally results in no performance improvement and may even be less performant due to the cost of context-switching.


## Configuration

Via environment variables:  

| Name                       | Default Value | Description                                                                                                                                                                                                                                                                                                                                                                                                      |
|----------------------------|:-------------:|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **GOTHON_KEEP_TEMP_DIR**   |    `false`    | If set to `true` (case-insensitive), the hidden `.gothon` directory that normally gets deleted after a run will remain.  This directory stores the Gothon-interpreted version of your project along with the collection of UDS socket files needed for IPC between Gothon and your script/application.  This is useful if you're getting unexpected results and suspect an issue with the Gothon-generated code. |
| **GOTHON_STRING_MAX_SIZE** |    `65536`    | The maximum size (in bytes) of the buffer used to store the text for a given `str` variable.  Exceeding this limit will produce unexpected results!                                                                                                                                                                                                                                                              |


Example:
```shell
export GOTHON_KEEP_TEMP_DIR=true; gothon 4 my_app
```


Via commented code in your Python module(s):

| Name                                | Default Value | Description                                                                                                                                         |
|-------------------------------------|:-------------:|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| **gothon:var_def:prefix**           |      `_`      | The case-sensitive string you must use as a prefix for all Gothon-managed variables you declare.  Set to `None` for no prefix requirement*.         |
| **gothon:var_def:suffix**           |      `_`      | The case-sensitive string you must use as a suffix for all Gothon-managed variables you declare.  Set to `None` for no suffix requirement*.         |
| **gothon:var_usage:require_parens** |    `False`    | If set to `True` / `true`, any time you _use_ (not _assign_ to) a variable, it must be encapsulated in parentheses**.                               |

<sub>*Setting both the prefix and suffix to `None` will likely result in generated code that is broken, unless your variable names are long/unique!</sub>

<sub>**This helps ensure Gothon is able to identify the variables it manages when parsing your code, however requiring parenthesis when not needed by a Python interpreter means your IDE will display warnings regarding redundant/unnecessary use of parenthesis.  In most cases, you do not need to enable this mode.

Example:
```python
# gothon:var_def:prefix = __
# gothon:var_def:suffix = None
# gothon:var_usage:require_parens = True

__my_gothon_managed_counter: int = 0
_my_python_managed_counter: int = 0

while (__my_gothon_managed_counter) < 100:  # will generate warning
    __my_gothon_managed_counter += 1
```

Note that all settings defined as comments within a module apply only to that module.

## Variables

Gothon has the notion of **system** and **user** variables:

  * **System** variables are automatically added to every module in your project that contains "Gothon code," such as configuration directives in commented code or the declaration of a Gothon-managed user variable.  Although you do not need to "forward-declare" these system variables before using them in a Gothon script, not doing so will result in errors reported by your IDE.  Regardless of what value you use to initialize a forward-declared system variable, when Gothon interprets your code it will initialize them using the correct values...however, you can but (likely) should not change them at runtime!  
  * **User** variables are those you define to hold and share your application's data/state across all node instances.  This type of variable also includes the synchronization primitives `lock`, `unlock`, and `sync` (as they are called in Gothon), all of which are of the Python type `callable`.  All other Gothon types (`bool`, `int`, `float`, `str`) map to Python types precisely.  
  
Note that both **system** and **user** variables respect the `prefix`/`suffix` configuration options!

System variables (the names shown assume the default prefix/suffix is used):

| Name           | Type  | Description                                                                                                                                        |
|----------------|:-----:|----------------------------------------------------------------------------------------------------------------------------------------------------|
| `_node_count_` | `int` | The number of instances of your script/application that will be started and managed by Gothon (the 1st argument you pass to the `gothon` command). |
| `_node_`       | `int` | The unique node ID assigned to the running instance of your script/application. Assigned ID's start at 0 and end at `_node_count_ - 1`.            |

Unless all nodes will be doing the same work (will follow the same workflow/algorithm), you will likely need to use these system variables.

To declare and use a Gothon-managed user variable, the following rules must be followed:
  1. The variable name must have the correct prefix and suffix, as specified by the `gothon:var_def:prefix`/`gothon:var_def:suffix` configuration options. By default, they are both `_`.
  2. If `gothon:var_usage:require_parens` is set to `true` (default is `false`), the variable name must be encapsulated in parentheses when used/referenced.
  3. Type hints must be used in the declaration of the variable (not necessary anywhere else).
  4. The type of the variable (as specified by the type hint) must be one of the supported types (see table below). 
  5. When declaring/initializing the variable, you may not use another Gothon-managed variable as part of the expression used to set the initial value. 

Although Gothon variables can be defined at the module level (global scope), within a function or control structure (local scope), or as a _class_ variable, the effective scope is always global to the module.  If you define a variable named `_x_` as a Python global variable and as a class variable in the same module, a Python interpreter recognizes them as two distinct variables but to Gothon they would reference the same variable in the memory it manages.  It's therefore vital that ALL Gothon variables, both system and user-defined, are uniquely named within a module, regardless of the namespace/scope to which they would normally belong if interpreted by CPython.

For example, while you may define a Gothon variable using either of these two approaches, doing both at the same time would not result in two unique variables in Gothon memory:
```python
_value_: int = 0

class Counter:
    _value_: int = 0
```

The reason for choosing one approach over another would then be based on normal design considerations...just avoid doing both so that the effect of running the code is clear.  Following this rule also helps ensure the script will produce the same effect when running normally via `python -m my_app` and when running Gothon with one node: `gothon 1 my_app`.

Supported Python types and the (atomic/concurrent) operations that can be performed on them:  

|           |        `=`         |        `+=`        |        `-=`        |        `*=`        |        `/=`        | **(any other operator)** |
|:---------:|:------------------:|:------------------:|:------------------:|:------------------:|:------------------:|:------------------------:|
| **bool**  | :heavy_check_mark: |        :x:         |        :x:         |        :x:         |        :x:         |           :x:            |
|  **int**  | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: |           :x:            |
| **float** | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: |           :x:            |
|  **str**  | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: |        :x:         |        :x:         |           :x:            |


### Concurrent Queue

Gothon also supports types `Queue[T]` and `LifoQueue[T]`.  These are modeled after the similarly named types defined in the `queue` and `multiprocessing` modules.  Because Gothon translates your code before feeding it to your Python interpreter, you do not need to import any modules before using these queue types, however if you prefer to suppress IDE warnings and benefit from autocomplete features, etc, then you can import either `queue` or `multiprocessing` when using Gothon's queue classes, as it too implements the same API (to a degree).  

Gothon's version of these classes are generic and expect you to pass in the type of the item the queue stores (`T`).  The type of `T` must be one of the four primitives Gothon supports (see table in previous section).  

Like with the other modules, you can pass in a maximum size for the queue to prevent it from growing beyond that limit.  If set to zero or not passed in the constructor, no limit will be enforced.

The following methods are available for `Queue[T]` / `LifoQueue[T]`:
  * `size() -> int` and `qsize() -> int`  Returns the number of items in the queue.
  * `empty() -> bool`  Returns `True` if there are no items in the queue.
  * `full() -> bool`  Returns `True` if the item count limit has been reached.
  * `put(val: T) -> bool`  Adds an item to the queue and returns `True` if the operation was successful (the queue was not full).
  * `get() -> (T, bool)`  Returns the next item or the default/empty value for that type if one was not available and `True` if one was available.

All calls are non-blocking and also note that the `put()` method returns a boolean to indicate success versus raising an exception like the other APIs.

Example:
```python
from queue import Queue

_node_: int = 0
_node_count_: int = 0

_sync_main_: callable = lambda n=_node_count_: ()

_numbers_: Queue[int] = Queue(10)
numbers = []

def add_number_until_full():
    current_number = 0
    while not _numbers_.full():
        ok = _numbers_.put(current_number)
        if ok:
            current_number += 1
    _sync_main_(1)


def remove_number_until_empty():
    while not _numbers_.empty():
        num, ok = _numbers_.get()
        if ok:
            numbers.append(num)


if __name__ == '__main__':
    add_number_until_full()
    _sync_main_()

    if _node_ == 0:
        print(f'size before empty: {_numbers_.qsize()}')
        remove_number_until_empty()
        print(f'size after empty: {_numbers_.qsize()}')
        numbers.sort()
        for number in numbers:
            print(f'{number} ')
```

### Synchronization Primitives

The Gothon types `lock`, `unlock`, and `sync` are only "types" in the conceptual sense...when defining them, use the Python type `callable`, then distinguish between them by using the right prefix when naming your variable (see table below).


| Gothon Type | Name Prefix |          Example Declaration          | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
|:-----------:|:-----------:|:-------------------------------------:|:--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
|  **lock**   |   `lock_`   |   `_lock_x_: callable = lambda: ()`   | Invoke this function to ensure only one node can execute the code that follows, until the matching `unlock` primitive is called from the node that invoked `lock`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| **unlock**  |  `unlock_`  |  `_unlock_x_: callable = lambda: ()`  | Invoke this function to signal to other nodes that the previously locked section of code can now be executed by another node.  The variable name must be the same as the `lock` primitive, excluding the **name prefix**.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
|  **sync**   |   `sync_`   | `_sync_x_: callable = lambda n=8: ()` | Use this primitive to ensure all nodes begin execution of a section of code at the same time or use it to have one node wait for others to complete their work before executing some code, etc.  When declaring the primitive, ensure the lambda function signature expects a single input parameter (can have any name) and set the default value to whatever you want the sync counter threshold to be.  When nodes invoke this function, they pass in an integer value that gets added to an internal counter...once that counter reaches the specified sync counter threshold, then whenever any node invokes the function without passing in a value it will return immediately, otherwise it will block until the internal sync counter reaches the threshold.  Nodes invoking the function passing in `1` usually do so to indicate they are done with their work. |

<sub>All examples assume the default variable `prefix`/`suffix` is used.  Also note that the synchronization primitive name prefix cannot be customized.</sub>

