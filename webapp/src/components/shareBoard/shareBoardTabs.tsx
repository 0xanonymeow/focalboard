// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {useState, useRef} from 'react'
import {useIntl, FormattedMessage} from 'react-intl'

import CompassIcon from '../../widgets/icons/compassIcon'
import BoardPermissionGate from '../permissions/boardPermissionGate'
import {Permission} from '../../constants'
import {useAppSelector} from '../../store/hooks'
import {getCurrentBoard} from '../../store/boards'
import {getMe} from '../../store/users'
import {useHasPermissions} from '../../hooks/permissions'

import EmailInvite from './emailInvite'
import PendingInvitations from './pendingInvitations'
import './shareBoardTabs.scss'

interface Props {
    children: React.ReactNode // This will be the existing share content (users, links, etc.)
}

const ShareBoardTabs = ({children}: Props) => {
    const intl = useIntl()
    const board = useAppSelector(getCurrentBoard)
    const me = useAppSelector(getMe)
    const canInvite = useHasPermissions(board.teamId, board.id, [Permission.ManageBoardRoles])
    
    const [activeTab, setActiveTab] = useState<'invite' | 'share'>(canInvite ? 'invite' : 'share')
    const pendingInvitationsRef = useRef<{refresh: () => void}>(null)

    const allTabs = [
        {
            id: 'invite' as const,
            label: intl.formatMessage({id: 'ShareBoard.tab-invite', defaultMessage: 'Invite People'}),
            icon: 'email-outline',
        },
        {
            id: 'share' as const,
            label: intl.formatMessage({id: 'ShareBoard.tab-share', defaultMessage: 'Share & Manage'}),
            icon: 'link-variant',
        },
    ]
    
    // Filter tabs based on permissions
    const tabs = canInvite ? allTabs : allTabs.filter(tab => tab.id !== 'invite')

    return (
        <div className='share-board-tabs'>
            <div className='share-board-tabs-header'>
                {tabs.map((tab) => (
                    <button
                        key={tab.id}
                        type='button'
                        className={`share-board-tab ${activeTab === tab.id ? 'active' : ''}`}
                        onClick={() => setActiveTab(tab.id)}
                    >
                        <CompassIcon icon={tab.icon}/>
                        <span>{tab.label}</span>
                    </button>
                ))}
            </div>

            <div className='share-board-tabs-content'>
                {activeTab === 'invite' && canInvite && (
                    <div className='invite-tab-content'>
                        <BoardPermissionGate permissions={[Permission.ManageBoardRoles]}>
                            <EmailInvite onInvitationSent={() => pendingInvitationsRef.current?.refresh()}/>
                            <PendingInvitations ref={pendingInvitationsRef}/>
                        </BoardPermissionGate>
                    </div>
                )}

                {activeTab === 'share' && (
                    <div className='share-tab-content'>
                        {children}
                    </div>
                )}
            </div>
        </div>
    )
}

export default ShareBoardTabs
