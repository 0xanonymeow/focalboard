// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {useState, useEffect} from 'react'
import {useIntl, FormattedMessage} from 'react-intl'

import {useAppSelector} from '../../store/hooks'
import {getCurrentBoard} from '../../store/boards'
import {Permission} from '../../constants'
import {useHasPermissions} from '../../hooks/permissions'

import client from '../../octoClient'
import IconButton from '../../widgets/buttons/iconButton'
import CompassIcon from '../../widgets/icons/compassIcon'
import {sendFlashMessage} from '../flashMessages'

import './pendingInvitations.scss'

interface BoardInvitation {
    id: string
    boardId: string
    email: string
    role: string
    createdBy: string
    createdAt: number
    expiresAt: number
    usedAt?: number
    usedBy?: string
    resendCooldownSeconds?: number
}

interface Props {
    onRefresh?: () => void
}

const PendingInvitations = React.memo(React.forwardRef<{refresh: () => void}, Props>(({onRefresh}, ref) => {
    const intl = useIntl()
    const board = useAppSelector(getCurrentBoard)
    const canManageRoles = useHasPermissions(board?.teamId, board?.id, [Permission.ManageBoardRoles])

    const [invitations, setInvitations] = useState<BoardInvitation[]>([])
    const [loading, setLoading] = useState(false)

    const loadInvitations = async () => {
        if (!board?.id || !canManageRoles) {
            return
        }

        setLoading(true)
        try {
            const pendingInvitations = await client.getBoardInvitations(board.id)

            // Filter out used invitations to show only pending ones
            const pending = pendingInvitations.filter((inv) => !inv.usedAt && new Date(inv.expiresAt * 1000) > new Date())
            
            setInvitations(pending)
        } catch (error) {
            // eslint-disable-next-line no-console
            console.error('Failed to load pending invitations:', error)
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => {
        loadInvitations()
    }, [board?.id, canManageRoles])

    // Update countdown timers every second
    useEffect(() => {
        const interval = setInterval(() => {
            setInvitations(prev => {
                const updated = prev.map(invitation => {
                    if (invitation.resendCooldownSeconds && invitation.resendCooldownSeconds > 0) {
                        return {
                            ...invitation,
                            resendCooldownSeconds: invitation.resendCooldownSeconds - 1
                        }
                    }
                    return invitation
                })
                return updated
            })
        }, 1000)
        
        return () => clearInterval(interval)
    }, [])

    // Expose refresh function to parent via ref
    React.useImperativeHandle(ref, () => ({
        refresh: loadInvitations
    }))

    const handleResendInvitation = async (invitation: BoardInvitation) => {
        try {
            const success = await client.resendInvitation(invitation.id)
            if (success) {
                sendFlashMessage({content: intl.formatMessage({id: 'ShareBoard.invitation-resent', defaultMessage: 'Invitation resent successfully'}), severity: 'normal'})
                // Refresh invitations to get updated cooldown from server
                loadInvitations()
            } else {
                sendFlashMessage({content: intl.formatMessage({id: 'ShareBoard.resend-error', defaultMessage: 'Failed to resend invitation'}), severity: 'high'})
            }
        } catch (error: any) {
            // eslint-disable-next-line no-console
            console.error('Failed to resend invitation:', error)
            
            // Check if it's a cooldown error from server
            if (error.message && error.message.includes('wait') && error.message.includes('seconds')) {
                sendFlashMessage({content: error.message, severity: 'high'})
            } else {
                sendFlashMessage({content: intl.formatMessage({id: 'ShareBoard.resend-error', defaultMessage: 'Failed to resend invitation'}), severity: 'high'})
            }
        }
    }

    const handleRemoveInvitation = async (invitation: BoardInvitation) => {
        try {
            const success = await client.deleteInvitation(invitation.id)
            if (success) {
                setInvitations((prev) => prev.filter((inv) => inv.id !== invitation.id))
                sendFlashMessage({content: intl.formatMessage({id: 'ShareBoard.invitation-removed', defaultMessage: 'Invitation removed successfully'}), severity: 'normal'})
                onRefresh?.()
            } else {
                sendFlashMessage({content: intl.formatMessage({id: 'ShareBoard.remove-error', defaultMessage: 'Failed to remove invitation'}), severity: 'high'})
            }
        } catch (error) {
            // eslint-disable-next-line no-console
            console.error('Failed to remove invitation:', error)
            sendFlashMessage({content: intl.formatMessage({id: 'ShareBoard.remove-error', defaultMessage: 'Failed to remove invitation'}), severity: 'high'})
        }
    }

    const formatDate = (timestamp: number) => {
        return new Date(timestamp * 1000).toLocaleDateString()
    }

    const getRoleDisplayName = (role: string) => {
        switch (role) {
        case 'admin':
            return intl.formatMessage({id: 'BoardMember.admin', defaultMessage: 'Admin'})
        case 'editor':
            return intl.formatMessage({id: 'BoardMember.editor', defaultMessage: 'Editor'})
        case 'commenter':
            return intl.formatMessage({id: 'BoardMember.commenter', defaultMessage: 'Commenter'})
        case 'viewer':
            return intl.formatMessage({id: 'BoardMember.viewer', defaultMessage: 'Viewer'})
        default:
            return role
        }
    }

    const getResendCooldownText = (invitation: BoardInvitation) => {
        if (!invitation.resendCooldownSeconds || invitation.resendCooldownSeconds <= 0) {
            return null
        }
        return `${invitation.resendCooldownSeconds}`
    }

    const isResendOnCooldown = (invitation: BoardInvitation) => {
        return invitation.resendCooldownSeconds && invitation.resendCooldownSeconds > 0
    }

    if (!canManageRoles) {
        return null
    }

    if (loading) {
        return (
            <div className='pending-invitations'>
                <div className='pending-invitations-header'>
                    Loading invitations...
                </div>
            </div>
        )
    }

    if (invitations.length === 0) {
        return (
            <div className='pending-invitations'>
                <div className='pending-invitations-header'>
                    <FormattedMessage
                        id='ShareBoard.no-pending-invitations'
                        defaultMessage='No pending invitations'
                    />
                </div>
            </div>
        )
    }

    return (
        <div className='pending-invitations'>
            <div className='pending-invitations-header'>
                <FormattedMessage
                    id='ShareBoard.pending-invitations'
                    defaultMessage='Pending Invitations'
                />
            </div>

            <div className='pending-invitations-list'>
                {invitations.map((invitation) => (
                    <div
                        key={invitation.id}
                        className='pending-invitation-row'
                    >
                        <div className='invitation-info'>
                            <div className='invitation-email'>
                                {invitation.email}
                            </div>
                            <div className='invitation-details'>
                                <span className='invitation-role'>
                                    {getRoleDisplayName(invitation.role)}
                                </span>
                                <span className='invitation-separator'>{'•'}</span>
                                <span className='invitation-status'>
                                    <FormattedMessage
                                        id='ShareBoard.pending'
                                        defaultMessage='Pending'
                                    />
                                </span>
                                <span className='invitation-separator'>{'•'}</span>
                                <span className='invitation-expires'>
                                    <FormattedMessage
                                        id='ShareBoard.expires'
                                        defaultMessage='Expires {date}'
                                        values={{date: formatDate(invitation.expiresAt)}}
                                    />
                                </span>
                            </div>
                        </div>

                        <div className='invitation-actions'>
                            <div className='resend-button-container'>
                                {isResendOnCooldown(invitation) ? (
                                    <div className='cooldown-button'>
                                        {getResendCooldownText(invitation)}
                                    </div>
                                ) : (
                                    <IconButton
                                        onClick={() => handleResendInvitation(invitation)}
                                        icon={<CompassIcon icon='email-outline'/>}
                                        title={intl.formatMessage({id: 'ShareBoard.resend-invitation', defaultMessage: 'Resend invitation'})}
                                        size='small'
                                    />
                                )}
                            </div>
                            <IconButton
                                onClick={() => handleRemoveInvitation(invitation)}
                                icon={<CompassIcon icon='trash-can-outline'/>}
                                title={intl.formatMessage({id: 'ShareBoard.remove-invitation', defaultMessage: 'Remove invitation'})}
                                size='small'
                            />
                        </div>
                    </div>
                ))}
            </div>
        </div>
    )
}))

PendingInvitations.displayName = 'PendingInvitations'

export default PendingInvitations