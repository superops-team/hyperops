load('shell.star', 'shell')

res_1 = shell.exec(dir="./", cmd="ls -l", timeout=5)
print(res_1)

res_3 = shell.exec("df -lh")
print(res_3)
res_2 = shell.exec(dir="./", cmd="pwd")
print(res_2)

res_5 = shell.exec(dir="./", cmd="pwdss")
print(res_5)
