# fcitx5-bamboo-mint

Bamboo Mint — Bộ gõ tiếng Việt độc lập cho Fcitx5, dựa trên [fcitx5-bamboo](https://github.com/fcitx/fcitx5-bamboo) và [bamboo-core](https://github.com/BambooEngine/bamboo-core).

## Mục đích

Dự án này tạo ra một addon độc lập tên **Bamboo Mint** (`bamboomint`), cho phép chạy **song song** với addon `bamboo` gốc trong cùng một phiên fcitx5. Nhờ đó, bạn có thể chuyển đổi qua lại giữa:

- **Bamboo Mint** (tiếng Việt)
- **Bamboo** gốc (tiếng Việt)
- **Mozc** (tiếng Nhật)
- **Pinyin** / **Cangjie** (tiếng Trung)
- **keyboard-us** (bàn phím tiếng Anh)

> **Lưu ý:** Đây là nhánh phát triển `dev`. Các kiểu gõ không phải Telex đang được loại bỏ dần để giảm độ phức tạp. Xem thư mục `requirement/` để biết chi tiết từng bước.

---

## Kiến trúc

```
fcitx5-bamboo/
├── bamboo/bamboo-core/        # Thư viện lõi Go (xử lý tiếng Việt)
├── src/
│   ├── bamboo.cpp              # Engine gốc (sẽ xóa sau)
│   └── mint/
│       ├── bamboomint.cpp      # Bamboo Mint — engine mới
│       ├── bamboomint.h
│       ├── bamboomintconfig.h
│       ├── CMakeLists.txt
│       ├── mint-addon.conf.in.in
│       └── mint.conf.in
├── data/
│   └── scalable/apps/          # Icon (bamboo đỏ, bamboo-mint xanh)
├── requirement/                 # Tài liệu điều tra và hướng dẫn
└── CMakeLists.txt
```

- **Go Core (`bamboo-core`)** — Xử lý logic tiếng Việt, build ra `bamboo-core.a` và `bamboo-core.h`.
- **C++ Engine (`src/mint/`)** — Wrapper fcitx5, build ra `libbamboomint.so`.

---

## Yêu cầu

- Fcitx5 ≥ 5.1.7
- CMake ≥ 3.6
- Go ≥ 1.22
- `extra-cmake-modules`
- `libfcitx5core-dev`, `libfcitx5config-dev`, `libfcitx5utils-dev`, `fcitx5-modules-dev`

### Cài gói hệ thống (Linux Mint / Ubuntu)

```bash
sudo apt install cmake extra-cmake-modules \
    libfcitx5core-dev libfcitx5config-dev libfcitx5utils-dev \
    fcitx5-modules-dev golang-go build-essential
```

---

## Build

```bash
# 1. Lấy submodule (nếu chưa có)
git submodule update --init --recursive

# 2. Tạo thư mục build
mkdir -p build && cd build
cmake -DCMAKE_EXPORT_COMPILE_COMMANDS=ON ..
make -j4
```

## Cài đặt

```bash
# Thư viện .so
sudo cp build/src/mint/libbamboomint.so /usr/lib/x86_64-linux-gnu/fcitx5/

# Cấu hình addon và input method
sudo cp build/src/mint/mint-addon.conf /usr/share/fcitx5/addon/bamboomint.conf
sudo cp build/src/mint/mint.conf /usr/share/fcitx5/inputmethod/mint.conf

# Từ điển
sudo cp data/vietnamese.cm.dict /usr/share/fcitx5/bamboo/

# Icon
sudo cp data/scalable/apps/fcitx_bamboo_mint.svg \
     /usr/share/icons/hicolor/scalable/apps/

# Khởi động lại fcitx5
fcitx5-remote -r
```

> Xem chi tiết đầy đủ trong [`requirement/phase1/REQ-002.md`](requirement/phase1/REQ-002.md).

---

## Các kiểu gõ hiện có

> Trạng thái: đang loại bỏ dần các kiểu không phải Telex.

| Kiểu gõ | Trạng thái |
|---------|-----------|
| Telex | ✅ Giữ |
| VNI | ⏳ Đang xóa (REQ-005) |
| VIQR | ⏳ Còn lại |
| Microsoft layout | ⏳ Còn lại |
| Telex 2 | ⏳ Còn lại |
| Telex + VNI | ⏳ Còn lại |
| Telex + VNI + VIQR | ⏳ Còn lại |
| Telex W | ⏳ Còn lại |
| VNI Bàn phím tiếng Pháp | ❌ Đã xóa (REQ-004) |

---

## Requirement

Tài liệu điều tra và hướng dẫn chi tiết nằm trong thư mục `requirement/`:

| Tệp | Nội dung |
|-----|---------|
| `REQ-001.md` | Tổng quan dự án Bamboo Mint — tạo addon độc lập |
| `REQ-002.md` | Hướng dẫn build và cài đặt (tiếng Việt) |
| `REQ-003.md` | Phân tích cấu trúc kiểu gõ toàn project |
| `REQ-004.md` | Điều tra xóa "VNI Bàn phím tiếng Pháp" |
| `REQ-005.md` | Điều tra xóa "VNI" |

---

## Giấy phép

LGPLv2.1+
