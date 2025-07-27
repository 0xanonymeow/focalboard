// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react'

import CompassIcon from './compassIcon'

import './loadingIcon.scss'

export default function LoadingIcon(): JSX.Element {
    return (
        <div className='LoadingIcon'>
            <CompassIcon icon='loading'/>
        </div>
    )
}