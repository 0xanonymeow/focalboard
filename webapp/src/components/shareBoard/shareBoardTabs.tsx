// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {useState, useRef} from 'react'
import {useIntl, FormattedMessage} from 'react-intl'

import CompassIcon from '../../widgets/icons/compassIcon'
import BoardPermissionGate from '../permissions/boardPermissionGate'
import {Permission} from '../../constants'

import EmailInvite from './emailInvite'
import PendingInvitations from './pendingInvitations'
import './shareBoardTabs.scss'

interface Props {
    children: React.ReactNode // This will be the existing share content (users, links, etc.)
}

const ShareBoardTabs = ({children}: Props) => {
    const intl = useIntl()
    const [activeTab, setActiveTab] = useState<'invite' | 'share'>('invite')
    const pendingInvitationsRef = useRef<{refresh: () => void}>(null)

    const tabs = [
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
                {activeTab === 'invite' && (
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