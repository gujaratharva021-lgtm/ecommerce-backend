import { useEffect, useState } from 'react'
import QRCode from 'qrcode'
import Layout from '../components/Layout'
import Modal from '../components/Modal'
import {
  listProducts,
  createProduct,
  updateProduct,
  deleteProduct,
  updateInventory,
  listCategories,
  uploadImage,
  generateProductBarcode,
  IMAGE_ORIGIN,
} from '../api/admin'
import type { Product, Category } from '../types/admin'

const emptyForm = {
  name: '',
  description: '',
  price: '',
  category_id: '',
  stock: '',
  image_url: '',
}

export default function Products() {
  const [products, setProducts] = useState<Product[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [showCreate, setShowCreate] = useState(false)
  const [editingProduct, setEditingProduct] = useState<Product | null>(null)

  const [form, setForm] = useState(emptyForm)
  const [isSaving, setIsSaving] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)
  const [isUploading, setIsUploading] = useState(false)
  const [generatingBarcodeId, setGeneratingBarcodeId] = useState<number | null>(null)

  async function load() {
    setIsLoading(true)
    setError(null)
    try {
      let allProducts: Product[] = []
      let page = 1
      while (true) {
        const res = await listProducts({ limit: 200, page })
        const batch = res.products ?? []
        allProducts = allProducts.concat(batch)
        if (batch.length === 0 || batch.length < 20) break
        page++
        if (page > 100) break
      }
      const categoriesRes = await listCategories()
      setProducts(allProducts)
      setCategories(categoriesRes.categories ?? categoriesRes ?? [])
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to load products.')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  function openCreate() {
    setForm(emptyForm)
    setFormError(null)
    setEditingProduct(null)
    setShowCreate(true)
  }

  function openEdit(p: Product) {
    setForm({
      name: p.name,
      description: p.description ?? '',
      price: String(p.price),
      category_id: String(p.category_id),
      stock: '',
      image_url: p.image_url ?? '',
    })
    setFormError(null)
    setEditingProduct(p)
    setShowCreate(true)
  }

  function closeModal() {
    setShowCreate(false)
    setEditingProduct(null)
  }

  async function handleFileSelect(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (!file) return
    setIsUploading(true)
    setFormError(null)
    try {
      const { image_url } = await uploadImage(file)
      setForm((f) => ({ ...f, image_url }))
    } catch (err: any) {
      setFormError(err.response?.data?.error ?? 'Failed to upload image.')
    } finally {
      setIsUploading(false)
      e.target.value = ''
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setFormError(null)

    if (!form.name.trim() || !form.price || !form.category_id) {
      setFormError('Name, price, and category are required.')
      return
    }

    setIsSaving(true)
    try {
      if (editingProduct) {
        await updateProduct(editingProduct.id, {
          name: form.name.trim(),
          description: form.description.trim(),
          price: parseFloat(form.price),
          category_id: parseInt(form.category_id, 10),
          image_url: form.image_url.trim(),
        })
      } else {
        await createProduct({
          name: form.name.trim(),
          description: form.description.trim(),
          price: parseFloat(form.price),
          category_id: parseInt(form.category_id, 10),
          image_url: form.image_url.trim(),
          stock: form.stock ? parseInt(form.stock, 10) : 0,
        })
      }
      closeModal()
      setForm(emptyForm)
      load()
    } catch (err: any) {
      setFormError(
        err.response?.data?.error ??
          `Failed to ${editingProduct ? 'update' : 'create'} product.`
      )
    } finally {
      setIsSaving(false)
    }
  }

  async function handleDelete(id: number) {
    if (!confirm('Delete this product? This cannot be undone.')) return
    try {
      await deleteProduct(id)
      setProducts((prev) => prev.filter((p) => p.id !== id))
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to delete product.')
    }
  }

  async function handleStockChange(id: number, newStock: number, warehouseId: number) {
    if (Number.isNaN(newStock) || newStock < 0) return
    try {
      await updateInventory(id, newStock, warehouseId)
      setProducts((prev) =>
        prev.map((p) =>
          p.id === id
            ? {
                ...p,
                inventories: p.inventories?.length
                  ? p.inventories.map((inv, i) =>
                      i === 0 ? { ...inv, stock: newStock } : inv
                    )
                  : [{ id: 0, warehouse_id: warehouseId, stock: newStock, in_stock: newStock > 0 }],
              }
            : p
        )
      )
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to update stock.')
    }
  }

  async function handleGenerateBarcode(product: Product) {
    setGeneratingBarcodeId(product.id)
    try {
      const { barcode } = await generateProductBarcode(product.id)
      setProducts((prev) => prev.map((p) => (p.id === product.id ? { ...p, barcode } : p)))
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to generate barcode.')
    } finally {
      setGeneratingBarcodeId(null)
    }
  }

  // Opens a small popup with a printable label: the product name, price,
  // the barcode as both a scannable QR code and its plain text value, then
  // triggers the browser print dialog. QR (not a 1D barcode image) because
  // it's what the warehouse app's camera scanner reliably decodes, and it's
  // trivial to render client-side with no extra rendering dependencies.
  async function handlePrintLabel(product: Product) {
    if (!product.barcode) return
    const qrDataUrl = await QRCode.toDataURL(product.barcode, { width: 220, margin: 1 })

    const win = window.open('', '_blank', 'width=320,height=420')
    if (!win) return
    win.document.write(`
      <html>
        <head>
          <title>Label - ${product.name}</title>
          <style>
            body { font-family: sans-serif; text-align: center; padding: 16px; }
            img { width: 180px; height: 180px; }
            h2 { font-size: 14px; margin: 8px 0 2px; }
            p { font-size: 12px; color: #444; margin: 0 0 8px; letter-spacing: 1px; }
          </style>
        </head>
        <body>
          <img src="${qrDataUrl}" alt="barcode" />
          <h2>${product.name}</h2>
          <p>${product.barcode}</p>
        </body>
      </html>
    `)
    win.document.close()
    win.focus()
    win.onload = () => win.print()
  }

  return (
    <Layout>
      <div className="p-8">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-xl font-semibold">Products</h1>
            <p className="text-sm text-slate-400 mt-1">
              {products.length} product{products.length !== 1 ? 's' : ''}
            </p>
          </div>
          <button
            onClick={openCreate}
            className="px-4 py-2 rounded-lg bg-indigo-500 hover:bg-indigo-400 text-white text-sm font-medium transition-colors"
          >
            + Add product
          </button>
        </div>

        {isLoading && <p className="text-slate-400">Loading...</p>}
        {error && <p className="text-red-400">{error}</p>}

        {!isLoading && !error && products.length === 0 && (
          <div className="border border-dashed border-slate-800 rounded-xl p-10 text-center text-slate-500">
            No products yet. Add your first one to get started.
          </div>
        )}

        {!isLoading && products.length > 0 && (
          <div className="border border-slate-800 rounded-xl overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-slate-900 text-slate-400 text-left">
                  <th className="px-4 py-3 font-medium">Image</th>
                  <th className="px-4 py-3 font-medium">Name</th>
                  <th className="px-4 py-3 font-medium">Price</th>
                  <th className="px-4 py-3 font-medium">Stock</th>
                  <th className="px-4 py-3 font-medium">Barcode</th>
                  <th className="px-4 py-3 font-medium"></th>
                </tr>
              </thead>
              <tbody>
                {products.map((p) => (
                  <tr key={p.id} className="border-t border-slate-800">
                    <td className="px-4 py-3">

                      {p.image_url ? (
                        <img
                          src={
                            p.image_url.startsWith('http')
                              ? p.image_url
                              : `${IMAGE_ORIGIN}${p.image_url}`
                          }
                          alt={p.name}
                          className="w-10 h-10 rounded-md object-cover border border-slate-700"
                        />
                      ) : (
                        <div className="w-10 h-10 rounded-md bg-slate-800 border border-slate-700" />
                      )}
                    </td>
                    <td className="px-4 py-3">{p.name}</td>
                    <td className="px-4 py-3">₹{p.price}</td>
                    <td className="px-4 py-3">
                      <input
                        type="number"
                        defaultValue={p.inventories?.[0]?.stock ?? 0}
                        min={0}
                        className="w-20 bg-slate-800 border border-slate-700 rounded-md px-2 py-1 text-sm"
                        onBlur={(e) =>
                          handleStockChange(
                            p.id,
                            parseInt(e.target.value, 10),
                            p.inventories?.[0]?.warehouse_id ?? 1
                          )
                        }
                      />
                    </td>
                    <td className="px-4 py-3">
                      {p.barcode ? (
                        <div className="flex items-center gap-2">
                          <span className="text-xs font-mono text-slate-300">{p.barcode}</span>
                          <button
                            onClick={() => handlePrintLabel(p)}
                            className="text-indigo-400 hover:text-indigo-300 text-xs"
                          >
                            Print
                          </button>
                        </div>
                      ) : (
                        <button
                          onClick={() => handleGenerateBarcode(p)}
                          disabled={generatingBarcodeId === p.id}
                          className="text-indigo-400 hover:text-indigo-300 text-xs disabled:opacity-40"
                        >
                          {generatingBarcodeId === p.id ? 'Generating...' : 'Generate'}
                        </button>
                      )}
                    </td>
                    <td className="px-4 py-3 text-right space-x-3">
                      <button
                        onClick={() => openEdit(p)}
                        className="text-indigo-400 hover:text-indigo-300 text-xs"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDelete(p.id)}
                        className="text-red-400 hover:text-red-300 text-xs"
                      >
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {showCreate && (
        <Modal
          title={editingProduct ? 'Edit product' : 'Add product'}
          onClose={closeModal}
        >
          <form onSubmit={handleSubmit} className="space-y-3">
            <div>
              <label className="text-xs text-slate-400 block mb-1">Name</label>
              <input
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </div>
            <div>
              <label className="text-xs text-slate-400 block mb-1">Description</label>
              <textarea
                value={form.description}
                onChange={(e) => setForm({ ...form, description: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
                rows={2}
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="text-xs text-slate-400 block mb-1">Price (₹)</label>
                <input
                  type="number"
                  value={form.price}
                  onChange={(e) => setForm({ ...form, price: e.target.value })}
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
                />
              </div>
              {!editingProduct && (
                <div>
                  <label className="text-xs text-slate-400 block mb-1">Stock</label>
                  <input
                    type="number"
                    value={form.stock}
                    onChange={(e) => setForm({ ...form, stock: e.target.value })}
                    className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
                  />
                </div>
              )}
            </div>
            <div>
              <label className="text-xs text-slate-400 block mb-1">Category</label>
              <select
                value={form.category_id}
                onChange={(e) => setForm({ ...form, category_id: e.target.value })}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              >
                <option value="">Select a category</option>
                {categories.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="text-xs text-slate-400 block mb-1">Image</label>
              <div className="flex items-center gap-3">
                {form.image_url && (
                  <img
                    src={
                      form.image_url.startsWith('http')
                        ? form.image_url
                        : `${IMAGE_ORIGIN}${form.image_url}`
                    }
                    alt="preview"
                    className="w-12 h-12 rounded-md object-cover border border-slate-700"
                  />
                )}
                <label className="flex-1">
                  <input
                    type="file"
                    accept="image/jpeg,image/jpg,image/png,image/webp"
                    onChange={handleFileSelect}
                    disabled={isUploading}
                    className="w-full text-xs text-slate-400 file:mr-3 file:py-2 file:px-3 file:rounded-lg file:border-0 file:bg-slate-800 file:text-slate-200 file:text-xs hover:file:bg-slate-700 file:cursor-pointer"
                  />
                  {isUploading && (
                    <p className="text-xs text-slate-400 mt-1">Uploading...</p>
                  )}
                </label>
              </div>
            </div>

            {formError && <p className="text-red-400 text-xs">{formError}</p>}

            <button
              type="submit"
              disabled={isSaving || isUploading}
              className="w-full py-2 rounded-lg bg-indigo-500 hover:bg-indigo-400 text-white text-sm font-medium transition-colors mt-2"
            >
              {isSaving
                ? 'Saving...'
                : editingProduct
                ? 'Save changes'
                : 'Create product'}
            </button>
          </form>
        </Modal>
      )}
    </Layout>
  )
}
