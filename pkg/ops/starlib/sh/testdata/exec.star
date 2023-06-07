load('shell.star', 'shell')

res_1 = shell.exec(dir="./", cmd="ls -l", timeout=5, cpu_limit_by_quota=50, memory_limit_by_mb=100)
print(res_1)

res_3 = shell.exec("df -lh", cpu_limit_by_quota=50, memory_limit_by_mb=100)
print(res_3)
res_2 = shell.exec(dir="./", cmd="pwd", cpu_limit_by_quota=50, memory_limit_by_mb=100)
print(res_2)

res_5 = shell.exec(dir="./", cmd="pwdss")
print(res_5)

