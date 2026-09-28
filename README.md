# KejarBill API

Backend KejarBill adalah layanan pengelolaan patungan, pembagian biaya, saldo antar-participant, dan settlement. Aplikasi ini ditujukan untuk situasi ketika seseorang membayar lebih dahulu, lalu biaya tersebut perlu dibebankan kepada beberapa orang dengan pembagian yang dapat dipertanggungjawabkan.

KejarBill mendukung participant berupa user terdaftar maupun guest, pembagian equal/custom/itemized, perhitungan kewajiban antar-orang, pencatatan settlement, serta payment method untuk menerima pembayaran.

## Gambaran Produk

Alur bisnis utama KejarBill:

```text
User mendaftar
    |
    v
User membuat atau bergabung ke group
    |
    v
Group memiliki registered participant dan guest participant
    |
    v
Member mencatat expense dan menentukan pembagian biaya
    |
    v
Backend membentuk account ledger dan menghitung saldo antar participant
    |
    v
Participant melakukan settlement atas kewajibannya
```

Contoh sederhana:

- Ronin membayar makan bersama sebesar Rp150.000.
- Biaya dibebankan kepada Ronin, Arthur, dan Asta.
- Backend menyimpan share aktual setiap participant.
- `account_ledger` mencatat siapa berutang kepada siapa.
- Ketika Arthur membayar Ronin, settlement membalik sebagian kewajiban tersebut.

Tujuan utama backend adalah menjaga agar angka expense, share, saldo, dan settlement tetap konsisten walaupun request diulang, terjadi pembayaran sebagian, atau data diakses bersamaan.

## Domain Utama

### User

User memiliki akun login, username, status akun, access token, dan refresh token. User dapat menjadi member di banyak group.

### Group

Group adalah ruang patungan. Group memiliki member dengan role:

- `owner`
- `admin`
- `member`

Member juga memiliki status lifecycle seperti `active`, `left`, dan `removed`.

### Group Participant

Participant adalah pihak yang ikut menanggung expense.

- `registered`: terhubung ke `users.user_id`.
- `guest`: tidak memiliki user login.

Guest dapat dibuat oleh owner/admin dan dapat diklaim menjadi participant registered melalui proses claim.

Participant ID adalah identitas finansial yang digunakan pada expense, ledger, dan settlement. Username hanya untuk kebutuhan display dan tidak digunakan sebagai foreign key transaksi.

### Expense

Expense menyimpan transaksi biaya yang dibayar oleh satu participant dan dibebankan kepada satu atau beberapa participant.

Tiga metode pembagian tersedia:

- `equal`: total dibagi rata kepada participant yang dipilih.
- `custom`: setiap participant menerima `share_amount` eksplisit.
- `itemized`: setiap item memiliki participant sendiri dan share item dihitung berdasarkan qty, unit price, dan participant item.

Expense yang dihapus menggunakan lifecycle soft-delete. Data finansial tidak dihapus secara fisik dari database.

### Account Ledger

`account_ledger` adalah sumber kebenaran finansial KejarBill. Balance tidak disimpan sebagai angka manual, tetapi diturunkan dari ledger.

Ledger memiliki arah:

```text
from_participant_id -> to_participant_id
```

Artinya participant `from` memiliki kewajiban kepada participant `to`.

Source ledger yang digunakan:

- `expense`: kewajiban yang berasal dari expense.
- `settlement`: pembalikan kewajiban karena pembayaran.
- `adjustment`: koreksi atau data finansial khusus.

Tidak ada kolom currency pada ledger. Nominal saat ini diperlakukan sebagai integer rupiah.

### Settlement

Settlement mencatat pembayaran antar participant. Settlement harus memvalidasi:

- sender dan recipient berbeda.
- sender memiliki hak mengirim atas nama participant tersebut.
- amount tidak melebihi outstanding balance.
- settlement dibuat sebagai `completed` setelah validasi berhasil.
- payment channel dan payment method sesuai.
- idempotency key wajib digunakan agar retry tidak membuat pembayaran ganda.

Settlement yang sudah completed dapat mengunci expense terkait agar histori pembayaran tidak menjadi ambigu.

### Payment Method

Payment method menyimpan informasi rekening atau metode penerimaan milik user.

- Account number disimpan dalam bentuk encrypted AES-256-GCM.
- List API hanya mengirim account number yang sudah dimask.
- Plaintext hanya tersedia melalui endpoint reveal yang memiliki authorization khusus.
- Visibility yang tersedia: `private`, `group_members`, dan `debtor_only`.

## Prinsip Finansial

Beberapa invariant penting yang harus tetap benar:

1. `sum(expense_participants.share_amount) == expenses.total_amount`.
2. Total share item selalu sama dengan `qty * unit_price`.
3. Participant payer tidak membuat kewajiban kepada dirinya sendiri.
4. Semua perubahan expense dan ledger dilakukan dalam satu transaction.
5. Settlement tidak boleh melebihi saldo outstanding.
6. Balance harus dapat direkonstruksi dari `account_ledger`.
7. Expense yang sudah memiliki settlement completed terkait tidak boleh diedit atau dihapus.
8. Equal split memberikan remainder satu rupiah kepada participant paling awal sesuai urutan request.
9. Itemized split menggunakan `expense_item_participants`, bukan kolom participant tunggal pada item.

## Arsitektur

Project menggunakan Go, Fiber, PostgreSQL, pgx, Redis, dan PASETO. Dependency injection dilakukan secara manual saat bootstrap.

```text
cmd/api/main.go
        |
        v
bootstrap.BuildDependency()
        |
        v
bootstrap.BuildApp() / RegisterRoute()
        |
        v
Auth middleware -> Handler -> Service -> Repository -> PostgreSQL
                                      |
                                      +-> Redis / security services
```

Layer memiliki tanggung jawab berikut:

- `handler`: parsing request, validasi input, mengambil user ID dari auth context, dan membentuk response.
- `service`: business rule, authorization, transaction boundary, perhitungan share, dan mapping domain.
- `repository`: SQL query dan row scanning. Repository tidak melakukan orchestration bisnis.
- `entity`: representasi domain/database.
- `dto`: bentuk request dan response API.
- `constants`: enum dan sentinel error domain.

Struktur module utama:

```text
internal/
  bootstrap/
  module/
    auth/
    user/
    group/
    group_member/
    group_participant/
    expense/
    ledger/
    settlement/
    payment_method/
    activity/
  shared/
    config/
    database/
    middleware/
    request/
    response/
    security/
```

## Keamanan

- Authentication menggunakan PASETO dan refresh session berbasis Redis.
- Password diproses melalui package security khusus.
- Account number payment method dienkripsi dengan AES-256-GCM.
- Authorization group selalu memeriksa active membership.
- Ownership payment method diverifikasi berdasarkan `payment_methods.user_id`, bukan participant row ID.
- Reveal payment method menggunakan nested group/participant/payment-method path untuk mencegah IDOR.
- Plaintext account number tidak disimpan ke database, cache, atau log.
- Reveal response menggunakan `no-store` dan `no-cache` headers.
- HTTP logging hanya mencatat metadata request, bukan body response.
- Mutasi settlement menggunakan idempotency key.

## Testing and Quality

Command utama:

```bash
go test ./...
go test ./internal/integration/ -v
go vet ./...
gofmt -w internal/
CGO_ENABLED=0 go build -mod=readonly ./...
```

Integration test membutuhkan PostgreSQL aktif dan `TEST_DATABASE_URL` sesuai konfigurasi test. Test integration mencakup expense split, ledger, settlement, payment method visibility/reveal, authorization, dan lifecycle expense.

`make test` masih menjalankan `mockery --all` sebagai compatibility command lama. Jika mockery tidak tersedia, jalankan `go test ./...` secara langsung.
