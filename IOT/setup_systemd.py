import paramiko

host = '192.168.50.216'
user = 'tristrac'
pw = 'OuroKronii314-'

service_content = """[Unit]
Description=Scarrow Field IoT Gateway Service
After=network.target bluetooth.target

[Service]
Type=simple
ExecStart=/home/tristrac/scarrow/scarrow-hub
WorkingDirectory=/home/tristrac/scarrow
User=root
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
"""

try:
    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    client.connect(host, 22, user, pw)

    def run_sudo(cmd):
        stdin, stdout, stderr = client.exec_command(f'sudo -S {cmd}')
        stdin.write(pw + '\n')
        stdin.flush()
        err = stderr.read().decode().strip()
        # Ignore the sudo password prompt in stderr
        if err and not err.startswith("[sudo]"):
            pass 
        return stdout.read().decode()

    print("1. Stopping tmux session...")
    client.exec_command('tmux kill-session -t scarrow_hub 2>/dev/null')

    print("2. Stopping old service...")
    run_sudo('systemctl stop scarrow.service')

    print("3. Writing new systemd service file...")
    sftp = client.open_sftp()
    with sftp.file('/tmp/scarrow.service', 'w') as f:
        f.write(service_content)
    sftp.close()
    
    run_sudo('mv /tmp/scarrow.service /etc/systemd/system/scarrow.service')
    run_sudo('chmod 644 /etc/systemd/system/scarrow.service')
    
    print("4. Reloading and starting service...")
    run_sudo('systemctl daemon-reload')
    run_sudo('systemctl enable scarrow.service')
    run_sudo('systemctl start scarrow.service')

    print("✅ Systemd service updated and started successfully!")

except Exception as e:
    print(f"Error: {e}")
finally:
    client.close()
