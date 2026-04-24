// productApi.ts — Client untuk fetch produk dari backend kita sendiri.
//
// Nota: Fungsi ini sedia ada untuk guna backend endpoint GET /api/items
// (protected route yang return produk dari database PostgreSQL kita).
//
// Pada masa ini StorePage fetch terus dari Xsolla Store API.
// Fungsi ni boleh digunakan kalau nak switch ke backend sebagai sumber data produk.

const API = import.meta.env.VITE_API_BASE_URL

// getProducts — fetch senarai produk dari backend dengan token auth
// Backend endpoint: GET /api/items (protected, perlukan JWT)
// File backend: internal/handler/product.go → GetItems()
export async function getProducts(token: string) {
  const res = await fetch(`${API}/api/items`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  })

  if (!res.ok) {
    throw new Error('Failed to fetch items')
  }

  return res.json()
}
