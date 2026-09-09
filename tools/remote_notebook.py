import sys
import paramiko

if sys.platform == "win32":
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")

def run(cmd):
    host = "192.168.1.106"
    user = "Frondabrick"
    password = "2201"

    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    try:
        client.connect(host, port=22, username=user, password=password, timeout=10)
        stdin, stdout, stderr = client.exec_command(cmd)
        out = stdout.read().decode("cp850", errors="replace").strip()
        err = stderr.read().decode("cp850", errors="replace").strip()
        if out:
            print(out)
        if err:
            print(f"[ERR] {err}", file=sys.stderr)
        client.close()
    except Exception as e:
        print(f"[-] Error conectando al notebook: {e}", file=sys.stderr)

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Uso: python tools/remote_notebook.py <comando>")
        sys.exit(1)
    run(" ".join(sys.argv[1:]))
