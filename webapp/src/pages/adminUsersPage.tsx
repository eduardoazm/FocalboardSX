// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {useState, useEffect, useMemo, useCallback} from 'react'
import {useHistory} from 'react-router-dom'

import Button from '../widgets/buttons/button'
import Dialog from '../components/dialog'
import FocalboardLogoIcon from '../widgets/icons/focalboard_logo'
import CompassIcon from '../widgets/icons/compassIcon'
import SearchIcon from '../widgets/icons/search'
import client from '../octoClient'
import {IUser} from '../user'
import {useAppSelector} from '../store/hooks'
import {getMe} from '../store/users'

import './adminUsersPage.scss'

const AdminUsersPage: React.FC = () => {
    const history = useHistory()
    const me = useAppSelector<IUser | null>(getMe)

    const [users, setUsers] = useState<IUser[]>([])
    const [loading, setLoading] = useState(true)
    const [searchTerm, setSearchTerm] = useState('')
    const [feedback, setFeedback] = useState<{type: 'success' | 'error', text: string} | null>(null)

    // Modal states
    const [showCreateModal, setShowCreateModal] = useState(false)
    const [createForm, setCreateForm] = useState({username: '', email: '', password: '', roles: 'member'})
    const [createError, setCreateError] = useState('')
    const [createSubmitting, setCreateSubmitting] = useState(false)

    const [editingUser, setEditingUser] = useState<IUser | null>(null)
    const [editForm, setEditForm] = useState({username: '', email: '', roles: 'member'})
    const [editError, setEditError] = useState('')
    const [editSubmitting, setEditSubmitting] = useState(false)

    const [resetUser, setResetUser] = useState<IUser | null>(null)
    const [resetForm, setResetForm] = useState({password: '', confirmPassword: ''})
    const [resetError, setResetError] = useState('')
    const [resetSubmitting, setResetSubmitting] = useState(false)

    const [deletingUser, setDeletingUser] = useState<IUser | null>(null)
    const [deleteSubmitting, setDeleteSubmitting] = useState(false)
    const [deleteError, setDeleteError] = useState('')

    const isAdmin = Boolean(
        me && (
            me.username === 'admin' ||
            me.roles?.includes('system_admin') ||
            me.permissions?.includes('manage_system')
        ),
    )

    const loadUsers = useCallback(async () => {
        setLoading(true)
        try {
            const list = await client.adminGetUsers()
            setUsers(list || [])
        } catch (err: any) {
            setFeedback({type: 'error', text: err?.message || 'Erro ao carregar lista de usuários.'})
        } finally {
            setLoading(false)
        }
    }, [])

    useEffect(() => {
        if (isAdmin) {
            loadUsers()
        }
    }, [isAdmin, loadUsers])

    const filteredUsers = useMemo(() => {
        if (!searchTerm.trim()) {
            return users
        }
        const term = searchTerm.toLowerCase().trim()
        return users.filter((u) =>
            u.username.toLowerCase().includes(term) ||
            (u.email && u.email.toLowerCase().includes(term)),
        )
    }, [users, searchTerm])

    // Handlers
    const handleOpenCreateModal = () => {
        setCreateForm({username: '', email: '', password: '', roles: 'member'})
        setCreateError('')
        setShowCreateModal(true)
    }

    const handleCreateSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setCreateError('')
        if (!createForm.username.trim()) {
            setCreateError('Nome de usuário é obrigatório.')
            return
        }
        if (!createForm.email.trim()) {
            setCreateError('E-mail é obrigatório.')
            return
        }
        if (createForm.password.length < 6) {
            setCreateError('A senha deve ter pelo menos 6 caracteres.')
            return
        }

        setCreateSubmitting(true)
        try {
            const res = await client.adminCreateUser({
                username: createForm.username.trim(),
                email: createForm.email.trim(),
                password: createForm.password,
                roles: createForm.roles,
            })
            if (res.code === 200) {
                setShowCreateModal(false)
                setFeedback({type: 'success', text: `Usuário '${createForm.username}' criado com sucesso!`})
                await loadUsers()
            } else {
                setCreateError(res.error || 'Erro ao cadastrar usuário.')
            }
        } catch (err: any) {
            setCreateError(err?.message || 'Falha de comunicação ao cadastrar usuário.')
        } finally {
            setCreateSubmitting(false)
        }
    }

    const handleOpenEditModal = (user: IUser) => {
        setEditingUser(user)
        setEditForm({
            username: user.username,
            email: user.email || '',
            roles: user.roles === 'system_admin' ? 'system_admin' : 'member',
        })
        setEditError('')
    }

    const handleEditSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        if (!editingUser) {
            return
        }
        setEditError('')

        setEditSubmitting(true)
        try {
            const res = await client.adminUpdateUser(editingUser.id, {
                username: editForm.username.trim(),
                email: editForm.email.trim(),
                roles: editForm.roles,
            })
            if (res.code === 200) {
                setEditingUser(null)
                setFeedback({type: 'success', text: `Usuário '${editingUser.username}' atualizado com sucesso!`})
                await loadUsers()
            } else {
                setEditError(res.error || 'Erro ao atualizar usuário.')
            }
        } catch (err: any) {
            setEditError(err?.message || 'Falha de comunicação ao atualizar usuário.')
        } finally {
            setEditSubmitting(false)
        }
    }

    const handleOpenResetModal = (user: IUser) => {
        setResetUser(user)
        setResetForm({password: '', confirmPassword: ''})
        setResetError('')
    }

    const handleResetSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        if (!resetUser) {
            return
        }
        setResetError('')
        if (resetForm.password.length < 6) {
            setResetError('A nova senha deve ter pelo menos 6 caracteres.')
            return
        }
        if (resetForm.password !== resetForm.confirmPassword) {
            setResetError('As senhas não coincidem.')
            return
        }

        setResetSubmitting(true)
        try {
            const res = await client.adminResetPassword(resetUser.id, resetForm.password)
            if (res.code === 200) {
                setResetUser(null)
                setFeedback({type: 'success', text: `Senha do usuário '${resetUser.username}' redefinida com sucesso!`})
            } else {
                setResetError(res.error || 'Erro ao redefinir senha.')
            }
        } catch (err: any) {
            setResetError(err?.message || 'Falha de comunicação ao redefinir senha.')
        } finally {
            setResetSubmitting(false)
        }
    }

    const handleOpenDeleteModal = (user: IUser) => {
        setDeletingUser(user)
        setDeleteError('')
    }

    const handleDeleteSubmit = async () => {
        if (!deletingUser) {
            return
        }
        setDeleteSubmitting(true)
        setDeleteError('')
        try {
            const res = await client.adminDeleteUser(deletingUser.id)
            if (res.code === 200) {
                setDeletingUser(null)
                setFeedback({type: 'success', text: `Usuário '${deletingUser.username}' excluído com sucesso!`})
                await loadUsers()
            } else {
                setDeleteError(res.error || 'Erro ao excluir usuário.')
            }
        } catch (err: any) {
            setDeleteError(err?.message || 'Falha de comunicação ao excluir usuário.')
        } finally {
            setDeleteSubmitting(false)
        }
    }

    // Access Denied View
    if (!isAdmin) {
        return (
            <div className='AdminUsersPage'>
                <div className='access-denied-wrapper'>
                    <div className='lock-icon-container'>
                        <CompassIcon icon='lock' />
                    </div>
                    <h2>Acesso Restrito</h2>
                    <p>Somente usuários com privilégio de administrador do sistema têm permissão para acessar o gerenciamento de usuários.</p>
                    <Button
                        filled={true}
                        emphasis='primary'
                        size='medium'
                        onClick={() => history.push('/')}
                    >
                        Voltar aos Quadros
                    </Button>
                </div>
            </div>
        )
    }

    return (
        <div className='AdminUsersPage'>
            {/* Header */}
            <header className='admin-header'>
                <div className='header-left'>
                    <div className='logo-icon'>
                        <FocalboardLogoIcon />
                    </div>
                    <div className='header-titles'>
                        <h1>Gerenciamento de Usuários</h1>
                        <p>Painel de controle e administração de usuários cadastrados</p>
                    </div>
                </div>
                <div className='header-right'>
                    <Button
                        emphasis='secondary'
                        size='medium'
                        onClick={() => history.push('/')}
                    >
                        ← Voltar aos Quadros
                    </Button>
                    <Button
                        filled={true}
                        emphasis='primary'
                        size='medium'
                        onClick={handleOpenCreateModal}
                    >
                        + Novo Usuário
                    </Button>
                </div>
            </header>

            {/* Main Content */}
            <main className='admin-content'>
                {feedback && (
                    <div className={`feedback-banner ${feedback.type}`}>
                        <span>{feedback.text}</span>
                        <button
                            type='button'
                            className='close-banner-btn'
                            onClick={() => setFeedback(null)}
                        >
                            ×
                        </button>
                    </div>
                )}

                {/* Search & Counter Controls */}
                <div className='admin-controls-card'>
                    <div className='search-box'>
                        <span className='search-icon'>
                            <SearchIcon />
                        </span>
                        <input
                            type='text'
                            placeholder='Buscar por usuário ou e-mail...'
                            value={searchTerm}
                            onChange={(e) => setSearchTerm(e.target.value)}
                        />
                    </div>
                    <div className='user-count-badge'>
                        Total de usuários: <strong>{filteredUsers.length}</strong> {searchTerm && `(de ${users.length})`}
                    </div>
                </div>

                {/* Users Table */}
                <div className='table-card'>
                    <table className='users-table'>
                        <thead>
                            <tr>
                                <th>Usuário</th>
                                <th>E-mail</th>
                                <th>Função</th>
                                <th>Data de Cadastro</th>
                                <th style={{textAlign: 'right'}}>Ações</th>
                            </tr>
                        </thead>
                        <tbody>
                            {loading ? (
                                <tr>
                                    <td colSpan={5} className='empty-row'>
                                        Carregando usuários...
                                    </td>
                                </tr>
                            ) : filteredUsers.length === 0 ? (
                                <tr>
                                    <td colSpan={5} className='empty-row'>
                                        {searchTerm ? 'Nenhum usuário corresponde aos critérios de busca.' : 'Nenhum usuário cadastrado.'}
                                    </td>
                                </tr>
                            ) : (
                                filteredUsers.map((u) => {
                                    const isCurrentUser = me?.id === u.id
                                    const isRootAdmin = u.username === 'admin'
                                    const isUserAdmin = u.roles === 'system_admin' || isRootAdmin
                                    const initial = (u.username || '?').charAt(0).toUpperCase()
                                    const formattedDate = u.create_at
                                        ? new Date(u.create_at).toLocaleDateString('pt-BR', {
                                            day: '2-digit',
                                            month: '2-digit',
                                            year: 'numeric',
                                        })
                                        : '-'

                                    return (
                                        <tr key={u.id}>
                                            <td>
                                                <div className='user-cell'>
                                                    <div className={`user-avatar ${isUserAdmin ? 'admin-avatar' : ''}`}>
                                                        {initial}
                                                    </div>
                                                    <div className='user-details'>
                                                        <div className='username-wrapper'>
                                                            <span className='username-text'>{u.username}</span>
                                                            {isCurrentUser && <span className='badge-you'>Você</span>}
                                                            {isRootAdmin && <span className='badge-root'>Root</span>}
                                                        </div>
                                                        <span className='user-id-sub'>{u.id}</span>
                                                    </div>
                                                </div>
                                            </td>
                                            <td className='email-cell'>
                                                {u.email || '-'}
                                            </td>
                                            <td>
                                                <span className={`role-badge ${isUserAdmin ? 'admin' : 'member'}`}>
                                                    {isUserAdmin ? 'Administrador' : 'Membro'}
                                                </span>
                                            </td>
                                            <td className='date-cell'>
                                                {formattedDate}
                                            </td>
                                            <td>
                                                <div className='actions-cell'>
                                                    <button
                                                        type='button'
                                                        className='btn-action'
                                                        title='Editar dados do usuário'
                                                        onClick={() => handleOpenEditModal(u)}
                                                    >
                                                        Editar
                                                    </button>
                                                    <button
                                                        type='button'
                                                        className='btn-action btn-reset'
                                                        title='Redefinir senha de acesso'
                                                        onClick={() => handleOpenResetModal(u)}
                                                    >
                                                        Nova Senha
                                                    </button>
                                                    <button
                                                        type='button'
                                                        className='btn-action btn-delete'
                                                        title={isCurrentUser ? 'Não é possível excluir o usuário conectado' : isRootAdmin ? 'Não é possível excluir o administrador padrão' : 'Excluir usuário'}
                                                        disabled={isCurrentUser || isRootAdmin}
                                                        onClick={() => handleOpenDeleteModal(u)}
                                                    >
                                                        Excluir
                                                    </button>
                                                </div>
                                            </td>
                                        </tr>
                                    )
                                })
                            )}
                        </tbody>
                    </table>
                </div>
            </main>

            {/* Modal: Novo Usuário */}
            {showCreateModal && (
                <Dialog
                    size='medium'
                    onClose={() => setShowCreateModal(false)}
                    title={<span>Cadastrar Novo Usuário</span>}
                >
                    <form className='admin-modal-content' onSubmit={handleCreateSubmit}>
                        <p className='modal-description'>
                            Preencha as informações para cadastrar uma nova conta de usuário no sistema Focalboard.
                        </p>

                        {createError && (
                            <div className='modal-alert-error'>
                                {createError}
                            </div>
                        )}

                        <div className='form-group'>
                            <label htmlFor='create-username'>Nome de Usuário *</label>
                            <input
                                id='create-username'
                                type='text'
                                required={true}
                                placeholder='ex: joaosilva'
                                value={createForm.username}
                                onChange={(e) => setCreateForm({...createForm, username: e.target.value})}
                            />
                        </div>

                        <div className='form-group'>
                            <label htmlFor='create-email'>E-mail *</label>
                            <input
                                id='create-email'
                                type='email'
                                required={true}
                                placeholder='ex: joao@empresa.com'
                                value={createForm.email}
                                onChange={(e) => setCreateForm({...createForm, email: e.target.value})}
                            />
                        </div>

                        <div className='form-group'>
                            <label htmlFor='create-password'>Senha Inicial *</label>
                            <input
                                id='create-password'
                                type='password'
                                required={true}
                                minLength={6}
                                placeholder='Mínimo de 6 caracteres'
                                value={createForm.password}
                                onChange={(e) => setCreateForm({...createForm, password: e.target.value})}
                            />
                            <div className='form-hint'>O usuário poderá alterar sua senha depois pelo menu do perfil.</div>
                        </div>

                        <div className='form-group'>
                            <label htmlFor='create-role'>Nível de Permissão *</label>
                            <select
                                id='create-role'
                                value={createForm.roles}
                                onChange={(e) => setCreateForm({...createForm, roles: e.target.value})}
                            >
                                <option value='member'>Membro (Acesso aos quadros atribuídos)</option>
                                <option value='system_admin'>Administrador do Sistema (Acesso total)</option>
                            </select>
                        </div>

                        <div className='modal-actions'>
                            <Button
                                emphasis='tertiary'
                                size='medium'
                                onClick={() => setShowCreateModal(false)}
                            >
                                Cancelar
                            </Button>
                            <Button
                                submit={true}
                                filled={true}
                                emphasis='primary'
                                size='medium'
                                disabled={createSubmitting}
                            >
                                {createSubmitting ? 'Cadastrando...' : 'Criar Usuário'}
                            </Button>
                        </div>
                    </form>
                </Dialog>
            )}

            {/* Modal: Editar Usuário */}
            {editingUser && (
                <Dialog
                    size='medium'
                    onClose={() => setEditingUser(null)}
                    title={<span>Editar Usuário: {editingUser.username}</span>}
                >
                    <form className='admin-modal-content' onSubmit={handleEditSubmit}>
                        <p className='modal-description'>
                            Altere as informações cadastrais e a função do usuário no sistema.
                        </p>

                        {editError && (
                            <div className='modal-alert-error'>
                                {editError}
                            </div>
                        )}

                        <div className='form-group'>
                            <label htmlFor='edit-username'>Nome de Usuário</label>
                            <input
                                id='edit-username'
                                type='text'
                                required={true}
                                disabled={editingUser.username === 'admin'}
                                value={editForm.username}
                                onChange={(e) => setEditForm({...editForm, username: e.target.value})}
                            />
                            {editingUser.username === 'admin' && (
                                <div className='form-hint'>O nome do usuário administrador raiz não pode ser alterado.</div>
                            )}
                        </div>

                        <div className='form-group'>
                            <label htmlFor='edit-email'>E-mail</label>
                            <input
                                id='edit-email'
                                type='email'
                                required={true}
                                value={editForm.email}
                                onChange={(e) => setEditForm({...editForm, email: e.target.value})}
                            />
                        </div>

                        <div className='form-group'>
                            <label htmlFor='edit-role'>Nível de Permissão</label>
                            <select
                                id='edit-role'
                                disabled={editingUser.username === 'admin'}
                                value={editForm.roles}
                                onChange={(e) => setEditForm({...editForm, roles: e.target.value})}
                            >
                                <option value='member'>Membro (Acesso aos quadros atribuídos)</option>
                                <option value='system_admin'>Administrador do Sistema (Acesso total)</option>
                            </select>
                            {editingUser.username === 'admin' && (
                                <div className='form-hint'>O usuário administrador raiz deve permanecer como Administrador.</div>
                            )}
                        </div>

                        <div className='modal-actions'>
                            <Button
                                emphasis='tertiary'
                                size='medium'
                                onClick={() => setEditingUser(null)}
                            >
                                Cancelar
                            </Button>
                            <Button
                                submit={true}
                                filled={true}
                                emphasis='primary'
                                size='medium'
                                disabled={editSubmitting}
                            >
                                {editSubmitting ? 'Salvando...' : 'Salvar Alterações'}
                            </Button>
                        </div>
                    </form>
                </Dialog>
            )}

            {/* Modal: Redefinir Senha */}
            {resetUser && (
                <Dialog
                    size='small'
                    onClose={() => setResetUser(null)}
                    title={<span>Redefinir Senha</span>}
                >
                    <form className='admin-modal-content' onSubmit={handleResetSubmit}>
                        <p className='modal-description'>
                            Defina uma nova senha de acesso para o usuário <strong>{resetUser.username}</strong>.
                        </p>

                        {resetError && (
                            <div className='modal-alert-error'>
                                {resetError}
                            </div>
                        )}

                        <div className='form-group'>
                            <label htmlFor='reset-password'>Nova Senha *</label>
                            <input
                                id='reset-password'
                                type='password'
                                required={true}
                                minLength={6}
                                placeholder='Mínimo de 6 caracteres'
                                value={resetForm.password}
                                onChange={(e) => setResetForm({...resetForm, password: e.target.value})}
                            />
                        </div>

                        <div className='form-group'>
                            <label htmlFor='reset-confirm-password'>Confirmar Nova Senha *</label>
                            <input
                                id='reset-confirm-password'
                                type='password'
                                required={true}
                                minLength={6}
                                placeholder='Repita a nova senha'
                                value={resetForm.confirmPassword}
                                onChange={(e) => setResetForm({...resetForm, confirmPassword: e.target.value})}
                            />
                        </div>

                        <div className='modal-actions'>
                            <Button
                                emphasis='tertiary'
                                size='medium'
                                onClick={() => setResetUser(null)}
                            >
                                Cancelar
                            </Button>
                            <Button
                                submit={true}
                                filled={true}
                                emphasis='primary'
                                size='medium'
                                disabled={resetSubmitting}
                            >
                                {resetSubmitting ? 'Alterando...' : 'Alterar Senha'}
                            </Button>
                        </div>
                    </form>
                </Dialog>
            )}

            {/* Modal: Excluir Usuário */}
            {deletingUser && (
                <Dialog
                    size='small'
                    onClose={() => setDeletingUser(null)}
                    title={<span>Confirmar Exclusão</span>}
                >
                    <div className='admin-modal-content'>
                        <p className='modal-description'>
                            Tem certeza que deseja excluir o usuário <strong>{deletingUser.username}</strong> ({deletingUser.email})?
                        </p>
                        <p style={{color: '#e53e3e', fontSize: '13px', marginTop: '-10px', marginBottom: '16px'}}>
                            ⚠️ Esta ação removerá o acesso desta conta ao sistema e encerrará todas as suas sessões ativas.
                        </p>

                        {deleteError && (
                            <div className='modal-alert-error'>
                                {deleteError}
                            </div>
                        )}

                        <div className='modal-actions'>
                            <Button
                                emphasis='tertiary'
                                size='medium'
                                onClick={() => setDeletingUser(null)}
                            >
                                Cancelar
                            </Button>
                            <Button
                                filled={true}
                                danger={true}
                                size='medium'
                                disabled={deleteSubmitting}
                                onClick={handleDeleteSubmit}
                            >
                                {deleteSubmitting ? 'Excluindo...' : 'Excluir Usuário'}
                            </Button>
                        </div>
                    </div>
                </Dialog>
            )}
        </div>
    )
}

export default React.memo(AdminUsersPage)
