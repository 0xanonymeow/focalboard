// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {createSlice, PayloadAction} from '@reduxjs/toolkit'

import {initialLoad, initialReadOnlyLoad} from './initialLoad'

import {RootState} from './index'

const globalErrorSlice = createSlice({
    name: 'globalError',
    initialState: {value: ''} as {value: string},
    reducers: {
        setGlobalError: (state, action: PayloadAction<string>) => {
            state.value = action.payload
        },
    },
    extraReducers: (builder) => {
        builder.addCase(initialReadOnlyLoad.rejected, (state, action) => {
            // Check if it's a network error or connection issue
            if (action.error.name === 'TypeError' || 
                action.error.message?.includes('fetch') ||
                action.error.message?.includes('Network') ||
                action.error.message?.includes('connection') ||
                !action.error.message) {
                state.value = 'network-error'
            } else {
                state.value = action.error.message || 'network-error'
            }
        })
        builder.addCase(initialLoad.rejected, (state, action) => {
            // Check if it's a network error or connection issue
            if (action.error.name === 'TypeError' || 
                action.error.message?.includes('fetch') ||
                action.error.message?.includes('Network') ||
                action.error.message?.includes('connection') ||
                !action.error.message) {
                state.value = 'network-error'
            } else {
                state.value = action.error.message || 'network-error'
            }
        })
    },
})

export const {setGlobalError} = globalErrorSlice.actions
export const {reducer} = globalErrorSlice

export const getGlobalError = (state: RootState): string => state.globalError.value
