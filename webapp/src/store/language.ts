// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import {createSlice, createAsyncThunk, PayloadAction} from '@reduxjs/toolkit'

import {getCurrentLanguage, storeLanguage as i18nStoreLanguage} from '../i18n'
import client from '../octoClient'
import {UserConfigPatch} from '../user'

import {RootState} from './index'
import {patchProps} from './users'

export const fetchLanguage = createAsyncThunk(
    'language/fetch',
    async () => getCurrentLanguage(),
)

export const storeLanguage = createAsyncThunk(
    'language/store',
    async (lang: string, {getState, dispatch}) => {
        i18nStoreLanguage(lang)
        const state = getState() as RootState
        const me = state.users?.me
        if (me?.id) {
            try {
                const patch: UserConfigPatch = {
                    updatedFields: {
                        language: lang,
                    },
                }
                const patchedProps = await client.patchUserConfig(me.id, patch)
                if (patchedProps) {
                    dispatch(patchProps(patchedProps))
                }
            } catch {
                // Ignore if offline
            }
        }
        return lang
    },
)

const languageSlice = createSlice({
    name: 'language',
    initialState: {value: 'pt-br'} as {value: string},
    reducers: {
        setLanguage: (state, action: PayloadAction<string>) => {
            state.value = action.payload
        },
    },
    extraReducers: (builder) => {
        builder.addCase(fetchLanguage.fulfilled, (state, action) => {
            state.value = action.payload
        })
        builder.addCase(storeLanguage.fulfilled, (state, action) => {
            state.value = action.payload
        })
    },
})

export const {setLanguage} = languageSlice.actions
export const {reducer} = languageSlice

export function getLanguage(state: RootState): string {
    return state.language.value
}
