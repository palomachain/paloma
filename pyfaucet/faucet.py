from bottle import post, request, run
import sqlite3
import subprocess

BANK = "paloma167rf0jmkkkqp9d4awa8rxw908muavqgghtw6tn"
AMOUNT = "100000grain"

db = sqlite3.connect("faucet.db")

@post("/claim")
def index():
    assert request.json["denom"] == "ugrain"
    address = request.json["address"]

    cur = db.cursor()
    if list(cur.execute("SELECT address FROM addresses WHERE address = ?", (address,))):
        return "funds previously sent, please wait a while"

    tx = subprocess.run([
        "palomad",
        "--node", "tcp://testnet.palomaswap.com:26657",
        "--fees", "200grain",
        "tx", "bank", "send", "-y",
        BANK,
        address,
        AMOUNT,
    ], capture_output=True)
    if tx.stderr:
        print(tx.stderr)
        return "failed"
    cur.execute("INSERT INTO addresses VALUES (?, date('now'))", (address,))
    db.commit()

    print(tx.stdout)
    return {}

run(host="localhost", port=8080)
