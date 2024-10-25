# GIF 转 PNG 工具

# 编译
go build -o gifconvert

# 转换为 PNG
./gifconvert -input example.gif -output ./output -format png

# 转换为 JPG
./gifconvert -input example.gif -output ./output -format jpg -quality 90

