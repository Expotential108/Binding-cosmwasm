package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageInfoHandlesMultipleCoins(t *testing.T) {
	info := MessageInfo{
		Sender: "foobar",
		Funds: []Coin{
			{Denom: "peth", Amount: "12345"},
			{Denom: "uatom", Amount: "789876"},
		},
	}
	bz, err := json.Marshal(info)
	require.NoError(t, err)

	// we can unmarshal it properly into struct
	var recover MessageInfo
	err = json.Unmarshal(bz, &recover)
	require.NoError(t, err)
	assert.Equal(t, info, recover)
}

func TestMessageInfoHandlesMissingCoins(t *testing.T) {
	info := MessageInfo{
		Sender: "baz",
	}
	bz, err := json.Marshal(info)
	require.NoError(t, err)

	// we can unmarshal it properly into struct
	var recover MessageInfo
	err = json.Unmarshal(bz, &recover)
	require.NoError(t, err)
	assert.Equal(t, info, recover)

	// make sure "funds":[] is in JSON
	var raw map[string]json.RawMessage
	err = json.Unmarshal(bz, &raw)
	require.NoError(t, err)
	funds, ok := raw["funds"]
	require.True(t, ok)
	assert.Equal(t, string(funds), "[]")
}

func TestBlockInfoSerialization(t *testing.T) {
	block := BlockInfo{
		Height:  123,
		Time:    1578939743_987654321,
		ChainID: "foobar",
	}
	bz, err := json.Marshal(block)
	require.NoError(t, err)
	assert.Equal(t, []byte(`{"height":123,"time":"1578939743987654321","chain_id":"foobar"}`), bz)

	block = BlockInfo{
		Height:  0,
		Time:    0,
		ChainID: "",
	}
	bz, err = json.Marshal(block)
	require.NoError(t, err)
	assert.Equal(t, []byte(`{"height":0,"time":"0","chain_id":""}`), bz)
}

func TestBlockInfoDeserialization(t *testing.T) {
	var block BlockInfo
	var err error

	// All set
	err = json.Unmarshal([]byte(`{"height":123,"time":"1578939743987654321","chain_id":"foobar"}`), &block)
	require.NoError(t, err)
	assert.Equal(t, BlockInfo{
		Height:  123,
		Time:    1578939743_987654321,
		ChainID: "foobar",
	}, block)

	// All zero
	err = json.Unmarshal([]byte(`{"height":0,"time":"0","chain_id":""}`), &block)
	require.NoError(t, err)
	assert.Equal(t, BlockInfo{
		Height:  0,
		Time:    0,
		ChainID: "",
	}, block)

	// All unset
	err = json.Unmarshal([]byte(`{}`), &block)
	require.NoError(t, err)
	assert.Equal(t, BlockInfo{
		Height:  0,
		Time:    0,
		ChainID: "",
	}, block)

	// All null 🤷‍
	err = json.Unmarshal([]byte(`{"height":null,"time":null,"chain_id":null}`), &block)
	require.NoError(t, err)
	assert.Equal(t, BlockInfo{
		Height:  0,
		Time:    0,
		ChainID: "",
	}, block)

	// Empty string is not a valid uint64 string
	err = json.Unmarshal([]byte(`{"time":""}`), &block)
	// TODO: Use ErrorContains once released (https://github.com/stretchr/testify/commit/6990a05d54)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid use of ,string struct tag, trying to unmarshal \"\" into uint64")
}
