# fcitx5-bamboo-mint

Bamboo Mint — Bộ gõ tiếng Việt độc lập cho Fcitx5, dựa trên [fcitx5-bamboo](https://github.com/fcitx/fcitx5-bamboo) và [bamboo-core](https://github.com/BambooEngine/bamboo-core).

## Mục đích

Dự án này tạo ra addon độc lập tên **Bamboo Mint** (`bamboomint`), chạy như một bộ gõ tiếng Việt riêng biệt trong fcitx5. Nhờ đó, bạn có thể chuyển đổi qua lại giữa:

- **Bamboo Mint** (tiếng Việt)
- **Mozc** (tiếng Nhật)
- **Pinyin** / **Cangjie** (tiếng Trung)
- **keyboard-us** (bàn phím tiếng Anh)

> **Lưu ý:** Đây là nhánh phát triển `dev`. Các kiểu gõ không phải Telex đã được loại bỏ trong Phase 1. Phase 3 đang điều tra các vấn đề còn lại. Xem thư mục `requirement/` để biết chi tiết từng bước.

---

## Kiến trúc

```
fcitx5-bamboo/
├── bamboo/bamboo-core/        # Thư viện lõi Go (xử lý tiếng Việt)
├── src/
│   └── mint/
│       ├── bamboomint.cpp      # Bamboo Mint — engine duy nhất
│       ├── bamboomint.h
│       ├── bamboomintconfig.h
│       ├── CMakeLists.txt
│       ├── mint-addon.conf.in.in
│       └── mint.conf.in
├── data/
│   └── scalable/apps/          # Icon bamboo-mint xanh
├── requirement/               # Tài liệu điều tra và hướng dẫn
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
# 1. Tạo thư mục build
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
sudo cp data/vietnamese.cm.dict /usr/share/fcitx5/bamboomint/

# Icon
sudo cp data/scalable/apps/fcitx_bamboo_mint.svg \
     /usr/share/icons/hicolor/scalable/apps/

# Khởi động lại fcitx5
fcitx5-remote -r
```

> Xem chi tiết đầy đủ trong [`requirement/phase1/REQ-002.md`](requirement/phase1/REQ-002.md).

---

## Các kiểu gõ hiện có

> Trạng thái: Phase 1 đã loại bỏ 6 kiểu gõ. Phase 3 đang đánh giá Telex 2 và Telex W.

| Kiểu gõ | Trạng thái |
|---------|-----------|
| Telex | ✅ Giữ |
| Telex 2 | ⏳ Phase 3 đánh giá |
| Telex W | ⏳ Phase 3 đánh giá |
| VNI | ❌ Đã xóa (REQ-005) |
| VIQR | ❌ Đã xóa (REQ-006) |
| Microsoft layout | ❌ Đã xóa (REQ-007) |
| Telex + VNI | ❌ Đã xóa (REQ-009) |
| Telex + VNI + VIQR | ❌ Đã xóa (REQ-008) |
| VNI Bàn phím tiếng Pháp | ❌ Đã xóa (REQ-004) |

---

## Requirement

Tài liệu điều tra và hướng dẫn chi tiết nằm trong thư mục `requirement/`:

### Phase 1 — Tạo addon độc lập và tinh gọn kiểu gõ

| Tệp | Nội dung |
|-----|---------|
| `phase1/REQ-001.md` | Tổng quan dự án Bamboo Mint — tạo addon độc lập |
| `phase1/REQ-002.md` | Hướng dẫn build và cài đặt (tiếng Việt) |
| `phase1/REQ-003.md` | Phân tích cấu trúc kiểu gõ toàn project |
| `phase1/REQ-004.md` | Điều tra xóa "VNI Bàn phím tiếng Pháp" |
| `phase1/REQ-005.md` | Điều tra xóa "VNI" |
| `phase1/REQ-006.md` | Xóa "VIQR" |
| `phase1/REQ-007.md` | Xóa "Microsoft layout" |
| `phase1/REQ-008.md` | Xóa "Telex + VNI + VIQR" |
| `phase1/REQ-009.md` | Xóa "Telex + VNI" |

### Phase 2 — Loại bỏ bamboo gốc, dọn dẹp project

| Tệp | Nội dung |
|-----|---------|
| `phase2/REQ-010.md` | Loại bỏ hoàn toàn engine `bamboo` gốc |
| `phase2/REQ-011.md` | Phân tích kiến trúc sau khi xóa bamboo gốc |
| `phase2/REQ-012.md` | Điều tra thay thế C++ bằng Rust |
| `phase2/REQ-013.md` | Điều tra thay thế UI fcitx5 |
| `phase2/REQ-014.md` | Phân tích bản quyền và nguồn gốc từng file |

### Phase 3 — Debug và cải thiện

| Tệp | Nội dung |
|-----|---------|
| `phase3/REQ-015.md` | Điều tra lỗi Zed Chat AI không kích hoạt Bamboo Mint |

---

## Giấy phép

LGPLv2.1+
