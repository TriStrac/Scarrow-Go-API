import paramiko
import os
import sys

host = "192.168.50.216"
port = 22
username = "tristrac"
password = "OuroKronii314-"

binary_path = os.path.abspath(r"pi-go-service\scarrow-hub")
remote_dir = "/home/tristrac/scarrow"
remote_bin = f"{remote_dir}/scarrow-hub"

print(f"Connecting to {host}...")
try:
    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    client.connect(host, port, username, password, timeout=10)
    print("Connected successfully.")

    # Create directory
    client.exec_command(f"mkdir -p {remote_dir}")

    # Upload binary
    print(f"Uploading {binary_path} to {remote_bin}...")
    sftp = client.open_sftp()
    sftp.put(binary_path, remote_bin)
    sftp.close()
    print("Upload complete.")

    # Make executable and run in tmux
    print("Setting up tmux session...")
    commands = [
        f"chmod +x {remote_bin}",
        "tmux kill-session -t scarrow_hub 2>/dev/null || true",
        f"cd {remote_dir} && tmux new-session -d -s scarrow_hub '{remote_bin}'"
    ]
    
    for cmd in commands:
        stdin, stdout, stderr = client.exec_command(cmd)
        err = stderr.read().decode().strip()
        if err and "no server running on" not in err:
            print(f"Warning/Error on '{cmd}': {err}")

    print("\n✅ Deployment successful!")
    print("To view the logs on the Pi, SSH in and run: tmux attach -t scarrow_hub")

except Exception as e:
    print(f"Deployment failed: {e}")
    sys.exit(1)
finally:
    client.close()
