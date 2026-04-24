// App.tsx — Komponen root yang pasang router ke dalam app.
// Semua navigation antara pages diuruskan oleh RouterProvider dari router.tsx.
// App sendiri takde logic — dia hanya jadi "pemegang" untuk router.

import { RouterProvider } from 'react-router-dom'
import { router } from './router'

export default function App() {
  return <RouterProvider router={router} />
}
