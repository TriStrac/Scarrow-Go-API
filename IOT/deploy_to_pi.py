import paramiko
import os

host = '192.168.50.216'
user = 'tristrac'
pw = 'OuroKronii314-'

def upload_dir(sftp, client, local_dir, remote_dir):
    try:
        sftp.mkdir(remote_dir)
    except:
        pass
    for item in os.listdir(local_dir):
        if item in ['.git', 'scarrow-hub']: continue
        local_path = os.path.join(local_dir, item)
        remote_path = remote_dir + '/' + item
        if os.path.isfile(local_path):
            print(f"Uploading {item}...")
            sftp.put(local_path, remote_path)
        elif os.path.isdir(local_path):
            upload_dir(sftp, client, local_path, remote_path)

try:
    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    client.connect(host, 22, user, pw)

    sftp = client.open_sftp()
    print("Uploading source code...")
    upload_dir(sftp, client, 'pi-go-service', '/home/tristrac/scarrow/src')
    sftp.close()

    print("Building on Pi (this ensures CGO works)...")
    stdin, stdout, stderr = client.exec_command('cd /home/tristrac/scarrow/src && /usr/local/go/bin/go build -o ../scarrow-hub main.go')
    exit_status = stdout.channel.recv_exit_status()
    if exit_status != 0:
        print(f"Build failed! \nError: {stderr.read().decode()}")
    else:
        print("Build successful. Starting tmux session...")
        client.exec_command('tmux kill-session -t scarrow_hub 2>/dev/null || true')
        client.exec_command('cd /home/tristrac/scarrow && tmux new-session -d -s scarrow_hub "./scarrow-hub"')
        print("\n✅ Deployment successful! Run 'tmux attach -t scarrow_hub' on the Pi.")

except Exception as e:
    print(f"Error: {e}")
finally:
    client.close()
