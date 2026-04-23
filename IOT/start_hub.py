import paramiko

host = '192.168.50.216'
user = 'tristrac'
pw = 'OuroKronii314-'

try:
    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    client.connect(host, 22, user, pw)

    print("Stopping old systemd service...")
    # Use -S to read password from stdin for sudo
    stdin, stdout, stderr = client.exec_command('sudo -S systemctl stop scarrow.service')
    stdin.write(pw + '\n')
    stdin.flush()
    
    print("Starting new Scarrow Hub in tmux...")
    client.exec_command('tmux kill-session -t scarrow_hub 2>/dev/null || true')
    # Run the binary and pipe output so we can see it in tmux
    client.exec_command('cd /home/tristrac/scarrow && tmux new-session -d -s scarrow_hub "./scarrow-hub"')

    print("\n✅ Success! The new binary is now running.")
    print("Run this on your Pi to see the logs: tmux attach -t scarrow_hub")

except Exception as e:
    print(f"Error: {e}")
finally:
    client.close()
