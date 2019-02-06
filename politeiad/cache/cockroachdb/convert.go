// Copyright (c) 2017-2019 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package cockroachdb

import (
	"strings"

	"github.com/decred/politeia/decredplugin"
	"github.com/decred/politeia/politeiad/cache"
)

func convertMDStreamFromCache(ms cache.MetadataStream) MetadataStream {
	return MetadataStream{
		ID:      ms.ID,
		Payload: ms.Payload,
	}
}

func convertRecordFromCache(r cache.Record) Record {
	metadata := make([]MetadataStream, 0, len(r.Metadata))
	for _, ms := range r.Metadata {
		metadata = append(metadata, convertMDStreamFromCache(ms))
	}

	files := make([]File, 0, len(r.Files))
	for _, f := range r.Files {
		files = append(files,
			File{
				Name:    f.Name,
				MIME:    f.MIME,
				Digest:  f.Digest,
				Payload: f.Payload,
			})
	}

	return Record{
		Key:       r.CensorshipRecord.Token + r.Version,
		Token:     r.CensorshipRecord.Token,
		Version:   r.Version,
		Status:    int(r.Status),
		Timestamp: r.Timestamp,
		Merkle:    r.CensorshipRecord.Merkle,
		Signature: r.CensorshipRecord.Signature,
		Metadata:  metadata,
		Files:     files,
	}
}

func convertRecordToCache(r Record) cache.Record {
	cr := cache.CensorshipRecord{
		Token:     r.Token,
		Merkle:    r.Merkle,
		Signature: r.Signature,
	}

	metadata := make([]cache.MetadataStream, 0, len(r.Metadata))
	for _, ms := range r.Metadata {
		metadata = append(metadata,
			cache.MetadataStream{
				ID:      ms.ID,
				Payload: ms.Payload,
			})
	}

	files := make([]cache.File, 0, len(r.Files))
	for _, f := range r.Files {
		files = append(files,
			cache.File{
				Name:    f.Name,
				MIME:    f.MIME,
				Digest:  f.Digest,
				Payload: f.Payload,
			})
	}

	return cache.Record{
		Version:          r.Version,
		Status:           cache.RecordStatusT(r.Status),
		Timestamp:        r.Timestamp,
		CensorshipRecord: cr,
		Metadata:         metadata,
		Files:            files,
	}
}

func convertCommentFromDecred(nc decredplugin.NewComment, ncr decredplugin.NewCommentReply) Comment {
	return Comment{
		Key:       nc.Token + ncr.CommentID,
		Token:     nc.Token,
		ParentID:  nc.ParentID,
		Comment:   nc.Comment,
		Signature: nc.Signature,
		PublicKey: nc.PublicKey,
		CommentID: ncr.CommentID,
		Receipt:   ncr.Receipt,
		Timestamp: ncr.Timestamp,
		Censored:  false,
	}
}

func convertCommentToDecred(c Comment) decredplugin.Comment {
	return decredplugin.Comment{
		Token:       c.Token,
		ParentID:    c.ParentID,
		Comment:     c.Comment,
		Signature:   c.Signature,
		PublicKey:   c.PublicKey,
		CommentID:   c.CommentID,
		Receipt:     c.Receipt,
		Timestamp:   c.Timestamp,
		TotalVotes:  0,
		ResultVotes: 0,
		Censored:    c.Censored,
	}
}

func convertLikeCommentFromDecred(lc decredplugin.LikeComment, lcr decredplugin.LikeCommentReply) LikeComment {
	return LikeComment{
		Token:     lc.Token,
		CommentID: lc.CommentID,
		Action:    lc.Action,
		Signature: lc.Signature,
		PublicKey: lc.PublicKey,
		Receipt:   lcr.Receipt,
		Timestamp: lcr.Timestamp,
	}
}

func convertLikeCommentToDecred(lc LikeComment) decredplugin.LikeComment {
	return decredplugin.LikeComment{
		Token:     lc.Token,
		CommentID: lc.CommentID,
		Action:    lc.Action,
		Signature: lc.Signature,
		PublicKey: lc.PublicKey,
		Receipt:   lc.Receipt,
		Timestamp: lc.Timestamp,
	}
}

func convertAuthorizeVoteToDecred(av AuthorizeVote) decredplugin.AuthorizeVote {
	return decredplugin.AuthorizeVote{
		Action:    av.Action,
		Token:     av.Token,
		Signature: av.Signature,
		PublicKey: av.PublicKey,
		Receipt:   av.Receipt,
		Timestamp: av.Timestamp,
	}
}

func convertStartVoteFromDecred(sv decredplugin.StartVote, svr decredplugin.StartVoteReply) StartVote {
	opts := make([]VoteOption, 0, len(sv.Vote.Options))
	for _, v := range sv.Vote.Options {
		opts = append(opts, VoteOption{
			Token:       sv.Vote.Token,
			ID:          v.Id,
			Description: v.Description,
			Bits:        v.Bits,
		})
	}
	return StartVote{
		Token:            sv.Vote.Token,
		Mask:             sv.Vote.Mask,
		Duration:         sv.Vote.Duration,
		QuorumPercentage: sv.Vote.QuorumPercentage,
		PassPercentage:   sv.Vote.PassPercentage,
		Options:          opts,
		PublicKey:        sv.PublicKey,
		Signature:        sv.Signature,
		StartBlockHeight: svr.StartBlockHeight,
		StartBlockHash:   svr.StartBlockHash,
		EndHeight:        svr.EndHeight,
		EligibleTickets:  strings.Join(svr.EligibleTickets, ","),
	}
}

func convertStartVoteToDecred(sv StartVote) (decredplugin.StartVote, decredplugin.StartVoteReply) {
	opts := make([]decredplugin.VoteOption, 0, len(sv.Options))
	for _, v := range sv.Options {
		opts = append(opts, decredplugin.VoteOption{
			Id:          v.ID,
			Description: v.Description,
			Bits:        v.Bits,
		})
	}

	dsv := decredplugin.StartVote{
		PublicKey: sv.PublicKey,
		Signature: sv.Signature,
		Vote: decredplugin.Vote{
			Token:            sv.Token,
			Mask:             sv.Mask,
			Duration:         sv.Duration,
			QuorumPercentage: sv.QuorumPercentage,
			PassPercentage:   sv.PassPercentage,
			Options:          opts,
		},
	}

	tix := strings.Split(sv.EligibleTickets, ",")
	dsvr := decredplugin.StartVoteReply{
		StartBlockHeight: sv.StartBlockHeight,
		StartBlockHash:   sv.StartBlockHash,
		EndHeight:        sv.EndHeight,
		EligibleTickets:  tix,
	}

	return dsv, dsvr
}

func convertCastVoteToDecred(cv CastVote) decredplugin.CastVote {
	return decredplugin.CastVote{
		Token:     cv.Token,
		Ticket:    cv.Ticket,
		VoteBit:   cv.VoteBit,
		Signature: cv.Signature,
	}
}