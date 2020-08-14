import websocket
import datetime
import time
import multiprocessing
import threading
import json


def receive(k, w, cnt):
    print("Receiver %d start at " % k, end='')
    print(datetime.datetime.utcnow())
    i = 0
    content = []
    while i < cnt:
        content.append(w.recv())
        i += 1
    print("Receiver %d end at " % k, end='')
    print(datetime.datetime.utcnow())
    w.close()
    print("Receiver %d total received message %d" % (k, len(content)))
    msg = {}
    for m in content:
        m = json.loads(m)
        if m["Sender"] not in msg:
            msg[m["Sender"]] = []
        msg[m["Sender"]].append(m["Index"])
    for m in msg:
        print(msg[m])
        for i in range(1, len(msg[m])):
            if msg[m][i] < msg[m][i - 1]:
                raise Exception("Incorrect order detected")


def send(k, i, cnt):
    ws = websocket.create_connection("ws://127.0.0.1:2808/%s" % k)
    print("Sender %d %d start at " % (k, i), end='')
    print(datetime.datetime.utcnow())
    for j in range(cnt):
        ws.send("{\"Type\": \"test\", \"Sender\": %d, \"Index\": %d}" % (i, j))
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

    p1 = multiprocessing.Process(target=test, args=(1, ))
    p2 = multiprocessing.Process(target=test, args=(2, ))
    p1.start()
    p2.start()
    p1.join()
    p2.join()

    # test(1)
