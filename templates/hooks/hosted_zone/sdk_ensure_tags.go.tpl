     r := rm.concreteResource(res)
     if r.ko == nil {
             // Should never happen... if it does, it's buggy code.
             panic("resource manager's EnsureTags method received resource with nil CR object")
     }
     defaultTags := ackrt.GetDefaultTags(&rm.cfg, r.ko, md)
     var existingTags []*svcapitypes.Tag
     existingTags = r.ko.Spec.Tags

     resourceTags, _ := convertToOrderedACKTags(existingTags)
     tags := acktags.Merge(resourceTags, defaultTags)
     r.ko.Spec.Tags = fromACKTags(tags, nil)
     return nil