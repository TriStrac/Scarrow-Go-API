import paramiko

host = '192.168.50.216'
user = 'tristrac'
pw = 'OuroKronii314-'

try:
    client = paramiko.SSHClient()
    client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
    client.connect(host, 22, user, pw)

    print("Upgrading Go on the Raspberry Pi to 1.23.2 (arm64)...")
    
    commands = [
        "wget https://go.dev/dl/go1.23.2.linux-arm64.tar.gz",
        "sudo rm -rf /usr/local/go",
        "sudo tar -C /usr/local -xzf go1.23.2.linux-arm64.tar.gz",
        "rm go1.23.2.linux-arm64.tar.gz"
    ]

    for cmd in commands:
        print(f"Executing: {cmd}")
        stdin, stdout, stderr = client.exec_command(f"sudo -S {cmd}" if "sudo" in cmd else cmd)
        if "sudo" in cmd:
            stdin.write(pw + '\n')
            stdin.flush()
        
        # Wait for command to finish
        exit_status = stdout.channel.recv_exit_status()
        if exit_status != 0:
            print(f"Command failed with exit status {exit_status}")
            print(stderr.read().decode())

    print("\n✅ Go upgrade complete. Verifying...")
    stdin, stdout, stderr = client.exec_command("/usr/local/go/bin/go version")
    print(stdout.read().decode())

except Exception as e:
    print(f"Error: {e}")
finally:
    client.close()
