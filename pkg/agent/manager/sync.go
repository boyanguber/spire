package manager

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"time"

	"github.com/spiffe/spire/pkg/agent/manager/cache"
	"github.com/spiffe/spire/pkg/common/bundleutil"
	"github.com/spiffe/spire/pkg/common/telemetry"
	"github.com/spiffe/spire/pkg/common/util"
	"github.com/spiffe/spire/proto/spire/api/node"
	"github.com/spiffe/spire/proto/spire/common"
)

type csrRequest struct {
	EntryID   string
	SpiffeID  string
	ExpiresAt time.Time
}

// synchronize hits the node api, checks for entries we haven't fetched yet, and fetches them.
func (m *manager) synchronize(ctx context.Context) (err error) {
	update, err := m.fetchUpdates(ctx, nil)
	if err != nil {
		return err
	}

	// update the cache and build a list of CSRs that need to be processed
	// in this interval.
	//
	// the values in `update` now belong to the cache. DO NOT MODIFY.
	var csrs []csrRequest
	var expiring int
	m.cache.Update(update, func(entry *common.RegistrationEntry, svid *cache.X509SVID) {
		var expiresAt time.Time
		switch {
		case svid == nil:
			// no SVID
		case len(svid.Chain) == 0:
			// SVID has an empty chain. this is not expected to happen.
		case isSVIDStale(m.c.Clk.Now(), svid.Chain[0]):
			// SVID has expired
			expiresAt = svid.Chain[0].NotAfter
			expiring++
		default:
			// SVID is good
			return
		}
		// we've exceeded the CSR limit, don't make any more CSRs
		if len(csrs) < node.CSRLimit {
			csrs = append(csrs, csrRequest{
				EntryID:   entry.EntryId,
				SpiffeID:  entry.SpiffeId,
				ExpiresAt: expiresAt,
			})
		}
	})
	m.c.Metrics.AddSample([]string{"cache_manager", "expiring_svids"}, float32(expiring))

	if len(csrs) > 0 {
		update, err := m.fetchUpdates(ctx, csrs)
		if err != nil {
			return err
		}
		// the values in `update` now belong to the cache. DO NOT MODIFY.
		m.cache.Update(update, nil)
	}
	return nil
}

func (m *manager) fetchUpdates(ctx context.Context, csrs []csrRequest) (_ *cache.CacheUpdate, err error) {
	counter := telemetry.StartCall(m.c.Metrics, "manager", "sync", "fetch_updates")
	defer counter.Done(&err)

	req := &node.FetchX509SVIDRequest{}

	// TODO: The node API is currently insufficient to handle multiple
	// registration entries mapping to the same SPIFFE ID (since results are
	// keyed by SPIFFE ID). This code will use the same SVID for any
	// registration entry sharing a SPIFFE ID and should be fixed when
	// the node API is fixed.
	privateKeys := make(map[string]*ecdsa.PrivateKey, len(csrs))
	for _, csr := range csrs {
		log := m.c.Log.WithField("spiffe_id", csr.SpiffeID)
		if !csr.ExpiresAt.IsZero() {
			log = log.WithField("expires_at", csr.ExpiresAt.Format(time.RFC3339))
		}
		log.Info("Renewing X509-SVID")
		counter.AddLabel("spiffe_id", csr.SpiffeID)

		// Skip CSR for the same SPIFFE ID... for now (see above TODO)
		if _, ok := privateKeys[csr.SpiffeID]; ok {
			continue
		}

		privateKey, csrBytes, err := newCSR(csr.SpiffeID)
		if err != nil {
			return nil, err
		}
		privateKeys[csr.SpiffeID] = privateKey
		req.Csrs = append(req.Csrs, csrBytes)
	}

	update, err := m.client.FetchUpdates(ctx, req)
	if err != nil {
		return nil, err
	}

	bundles, err := parseBundles(update.Bundles)
	if err != nil {
		return nil, err
	}

	bySpiffeID := make(map[string]*cache.X509SVID, len(update.SVIDs))
	for spiffeID, svid := range update.SVIDs {
		privateKey, ok := privateKeys[spiffeID]
		if !ok {
			continue
		}
		chain, err := x509.ParseCertificates(svid.CertChain)
		if err != nil {
			return nil, err
		}
		bySpiffeID[spiffeID] = &cache.X509SVID{
			Chain:      chain,
			PrivateKey: privateKey,
		}
	}

	byEntryID := make(map[string]*cache.X509SVID, len(bySpiffeID))
	for _, entry := range update.Entries {
		svid, ok := bySpiffeID[entry.SpiffeId]
		if !ok {
			continue
		}
		byEntryID[entry.EntryId] = svid
	}

	return &cache.CacheUpdate{
		Bundles:             bundles,
		RegistrationEntries: update.Entries,
		X509SVIDs:           byEntryID,
	}, nil
}

func newCSR(spiffeID string) (pk *ecdsa.PrivateKey, csr []byte, err error) {
	pk, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return
	}
	csr, err = util.MakeCSR(pk, spiffeID)
	if err != nil {
		return nil, nil, err
	}
	return
}

func parseBundles(bundles map[string]*common.Bundle) (map[string]*cache.Bundle, error) {
	out := make(map[string]*cache.Bundle, len(bundles))
	for _, bundle := range bundles {
		bundle, err := bundleutil.BundleFromProto(bundle)
		if err != nil {
			return nil, err
		}
		out[bundle.TrustDomainID()] = bundle
	}
	return out, nil
}

func isSVIDStale(now time.Time, svid *x509.Certificate) bool {
	ttl := svid.NotAfter.Sub(now)
	lifetime := svid.NotAfter.Sub(svid.NotBefore)
	return ttl < lifetime/2
}
