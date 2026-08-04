import { useEffect, useState } from 'react'
import Layout from '../components/Layout'
import Modal from '../components/Modal'
import { listCategories, createCategory, updateCategory, deleteCategory } from '../api/admin'
import type { Category } from '../types/admin'

export default function Categories() {
  const [categories, setCategories] = useState<Category[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [newName, setNewName] = useState('')
  const [isSaving, setIsSaving] = useState(false)

  const [editingCategory, setEditingCategory] = useState<Category | null>(null)
  const [editName, setEditName] = useState('')
  const [editError, setEditError] = useState<string | null>(null)
  const [isEditSaving, setIsEditSaving] = useState(false)

  async function load() {
    setIsLoading(true)
    setError(null)
    try {
      const res = await listCategories()
      setCategories(res.categories ?? res ?? [])
    } catch (err: any) {
      setError(err.response?.data?.error ?? 'Failed to load categories.')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault()
    if (!newName.trim()) return
    setIsSaving(true)
    try {
      await createCategory(newName.trim())
      setNewName('')
      load()
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to create category.')
    } finally {
      setIsSaving(false)
    }
  }

  function openEdit(c: Category) {
    setEditingCategory(c)
    setEditName(c.name)
    setEditError(null)
  }

  async function handleEditSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!editingCategory) return
    if (!editName.trim()) {
      setEditError('Category name is required.')
      return
    }
    setIsEditSaving(true)
    try {
      await updateCategory(editingCategory.id, editName.trim())
      setEditingCategory(null)
      load()
    } catch (err: any) {
      setEditError(err.response?.data?.error ?? 'Failed to update category.')
    } finally {
      setIsEditSaving(false)
    }
  }

  async function handleDelete(id: number) {
    if (!confirm('Delete this category?')) return
    try {
      await deleteCategory(id)
      setCategories((prev) => prev.filter((c) => c.id !== id))
    } catch (err: any) {
      alert(err.response?.data?.error ?? 'Failed to delete category.')
    }
  }

  return (
    <Layout>
      <div className="p-8 max-w-xl">
        <h1 className="text-xl font-semibold mb-1">Categories</h1>
        <p className="text-sm text-slate-400 mb-6">
          {categories.length} categor{categories.length !== 1 ? 'ies' : 'y'}
        </p>
        <form onSubmit={handleAdd} className="flex gap-2 mb-6">
          <input
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            placeholder="New category name"
            className="flex-1 bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
          />
          <button
            type="submit"
            disabled={isSaving}
            className="px-4 py-2 rounded-lg bg-indigo-500 hover:bg-indigo-400 text-white text-sm font-medium transition-colors"
          >
            Add
          </button>
        </form>
        {isLoading && <p className="text-slate-400">Loading...</p>}
        {error && <p className="text-red-400">{error}</p>}
        {!isLoading && categories.length === 0 && (
          <div className="border border-dashed border-slate-800 rounded-xl p-10 text-center text-slate-500">
            No categories yet.
          </div>
        )}
        <div className="space-y-2">
          {categories.map((c) => (
            <div
              key={c.id}
              className="flex items-center justify-between border border-slate-800 rounded-lg px-4 py-3"
            >
              <span className="text-sm">{c.name}</span>
              <div className="space-x-3">
                <button
                  onClick={() => openEdit(c)}
                  className="text-indigo-400 hover:text-indigo-300 text-xs"
                >
                  Edit
                </button>
                <button
                  onClick={() => handleDelete(c.id)}
                  className="text-red-400 hover:text-red-300 text-xs"
                >
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>

      {editingCategory && (
        <Modal title="Edit category" onClose={() => setEditingCategory(null)}>
          <form onSubmit={handleEditSubmit} className="space-y-3">
            <div>
              <label className="text-xs text-slate-400 block mb-1">Category name</label>
              <input
                value={editName}
                onChange={(e) => setEditName(e.target.value)}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-sm"
              />
            </div>

            {editError && <p className="text-red-400 text-xs">{editError}</p>}

            <button
              type="submit"
              disabled={isEditSaving}
              className="w-full py-2 rounded-lg bg-indigo-500 hover:bg-indigo-400 text-white text-sm font-medium transition-colors mt-2"
            >
              {isEditSaving ? 'Saving...' : 'Save changes'}
            </button>
          </form>
        </Modal>
      )}
    </Layout>
  )
}
