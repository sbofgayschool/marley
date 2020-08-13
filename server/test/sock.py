import websocket
import datetime
import time
import multiprocessing
import threading


def receive(k, w, cnt):
    print("Receiver %d start at " % k, end='')
    print(datetime.datetime.utcnow())
    i = 0
    while i < cnt:
        if w.recv() != "{\"Content\":\"pong\"}\n":
            raise Exception("Incorrect content")
        i += 1
    print("Receiver %d end at " % k, end='')
    print(datetime.datetime.utcnow())
    w.close()


def send(k, i, cnt):
    ws = websocket.create_connection("ws://127.0.0.1:2808/%s" % k)
    print("Sender %d %d start at " % (k, i), end='')
    print(datetime.datetime.utcnow())
    for j in range(cnt):
        ws.send("{\"Type\": \"test\", \"Content\": \"ping\"}")
        time.sleep(0.01)
    print("Sender %d %d end at " % (k, i), end='')
    print(datetime.datetime.utcnow())
    ws.close()


def test(k):
    num = 1000
    process = 20
    t = threading.Thread(
        target=receive, args=(k, websocket.create_connection("ws://127.0.0.1:2808/%s" % k), num * process)
    )
    t.start()
    processes = []
    for i in range(process):
        p = multiprocessing.Process(target=send, args=(k, i, num))
        p.start()
        processes.append(p)
    for p in processes:
        p.join()
    t.join()


if __name__ == "__main__":
    # websocket.enableTrace(True)
    '''
    p1 = multiprocessing.Process(target=test, args=(1, ))
    p2 = multiprocessing.Process(target=test, args=(2, ))
    p1.start()
    p2.start()
    p1.join()
    p2.join()
    '''
    test(1)
